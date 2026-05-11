# --- Fase 1: Build ---
FROM golang:1.22-alpine AS builder

# Instalamos dependencias necesarias para la compilación
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Aprovechar el caché de Docker para las dependencias
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilación estática: 
# CGO_ENABLED=0 para que no dependa de librerías dinámicas del host
# -ldflags="-s -w" para eliminar símbolos de debug y reducir tamaño
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o vault-api ./main.go

# --- Fase 2: Producción (Runtime) ---
FROM alpine:latest

# Añadir certificados para que la API pueda hablar con otros servicios vía HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Crear un usuario no privilegiado
RUN adduser -D -g '' vaultuser

WORKDIR /app

# Copiamos solo el binario compilado desde la fase de builder
COPY --from=builder /app/vault-api .

# Cambiamos el dueño de la carpeta al usuario no-root
RUN chown -R vaultuser:vaultuser /app

# Usar el usuario no-root
USER vaultuser

# Exponemos el puerto de la API
EXPOSE 8080

CMD ["./vault-api"]
