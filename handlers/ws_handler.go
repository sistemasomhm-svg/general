package handlers

import (
	"net/http"
	"github.com/gorilla/websocket"
	appws "vault-backend/websocket" 
)

// We assume we can inject the hub here or define it. For simplicity matching the snippet, we can add it to a Config/App struct or pass it.
// To keep it clean and compilable:

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// En producción, validar contra dominios permitidos para evitar CSWH
		return true 
	},
}

var AppHub = appws.NewHub()

func ServeWS(w http.ResponseWriter, r *http.Request) {
	// 1. Extraer UserID del contexto (puesto por nuestro AuthMiddleware)
    // Assuming context value is string
	userID, ok := r.Context().Value("user_id").(string)
    if !ok || userID == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	// 2. Upgrade de la conexión
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	AppHub.Register(userID, conn)

	// 3. Mantener conexión viva y manejar cierre
	go func() {
		defer func() {
			AppHub.Unregister(userID, conn)
			conn.Close()
		}()

		for {
			// Escuchamos por si el cliente envía algo o para detectar desconexión (Ping/Pong)
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
