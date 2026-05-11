package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vault-backend/database"
	"vault-backend/handlers"
	"vault-backend/middleware"
	"vault-backend/repositories"
	"vault-backend/services"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Configurar Logger Estructurado (slog)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 2. Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		slog.Info("Iniciando en modo producción (variables del sistema)")
	} else {
		slog.Info("Iniciando en modo desarrollo (.env cargado)")
	}

	// 3. Conectar a PostgreSQL (pgx/v5)
	database.ConnectDB()
	defer database.Pool.Close()

	// 4. Inicializar Repositorios, Servicios y Handlers (DI)
	userRepo := repositories.NewUserRepository(database.Pool)
	entryRepo := repositories.NewEntryRepository(database.Pool)

	authService := services.NewAuthService(userRepo)
	syncService := services.NewSyncService(entryRepo)

	authHandler := handlers.NewAuthHandler(authService)
	syncHandler := handlers.NewSyncHandler(syncService)

	// 5. Configurar Rutas
	mux := http.NewServeMux()

	// Rutas públicas
	mux.HandleFunc("/api/login", authHandler.Login)
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/salt", authHandler.GetSalt)

	// Rutas protegidas (Requieren JWT)
	mux.Handle("/api/sync", middleware.AuthMiddleware(http.HandlerFunc(syncHandler.Sync)))

	// WebSockets
	mux.Handle("/api/ws", middleware.AuthMiddleware(http.HandlerFunc(handlers.ServeWS)))

	// Health Check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 6. Iniciar Servidor con Graceful Shutdown
	port := os.Getenv("PORT") // Render usa esta variable
	if port == "" {
		port = os.Getenv("API_PORT") // Fallback a variable local
	}
	if port == "" {
		port = "8080" // Puerto estándar por defecto
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal para escuchar señales del SO
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Servidor iniciado", "puerto", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error iniciando el servidor", "error", err)
			os.Exit(1)
		}
	}()

	// Esperar señal de parada
	<-stop
	slog.Info("Apagando servidor...")

	// Timeout de 10 segundos para cerrar conexiones pendientes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error durante el apagado", "error", err)
	}

	slog.Info("Servidor detenido correctamente")
}
