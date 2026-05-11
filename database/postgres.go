package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URL") // Estándar en Render/managed DBs
	if dsn == "" {
		dsn = os.Getenv("DB_URL") // Fallback a variable personalizada
	}

	if dsn == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")
		if host != "" && user != "" && pass != "" && name != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
		} else {
			log.Fatal("DATABASE_URL, DB_URL o variables de base de datos requeridas")
		}
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Error parseando la configuración de DB: %v", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Error conectando a PostgreSQL (pgx): %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("No se pudo hacer ping a la base de datos: %v", err)
	}

	Pool = pool
	log.Println("Conexión a PostgreSQL (pgx/v5) establecida exitosamente!")
}
