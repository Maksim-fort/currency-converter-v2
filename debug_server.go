package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	// Простой тестовый сервер
	addr := "0.0.0.0:8080"

	fmt.Printf("Trying to listen on: %s\n", addr)

	// 1. Пробуем создать listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("❌ Failed to create listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("✅ Listener created successfully on %s\n", addr)

	// 2. Настраиваем HTTP сервер
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"healthy"}`)
	})

	fmt.Println("🚀 Starting HTTP server...")

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 3. Запускаем сервер
	if err := server.Serve(listener); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
