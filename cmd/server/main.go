package main

import (
	"currency-converter-v2/internal/app"
	"currency-converter-v2/internal/config"
	"log"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Создаем приложение
	app := app.New(cfg)

	// Просто запускаем
	log.Println("Starting Currency Converter API...")
	log.Println("🌐 API доступен: http://localhost:" + cfg.Server.Port)
	log.Println("💰 Фронтенд доступен: http://localhost:" + cfg.Server.Port + "/ui")

	if err := app.Run(); err != nil {
		log.Fatalf("Failed: %v", err)
	}

	log.Println("Stopped")
}
