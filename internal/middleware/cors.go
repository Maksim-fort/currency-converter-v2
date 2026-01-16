package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware создает middleware для CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Разрешаем запросы с любого источника
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		// Разрешаем методы
		c.Writer.Header().Set("Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// Разрешаем заголовки
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Origin, Content-Type, Content-Length, Accept-Encoding, "+
				"X-CSRF-Token, Authorization, Accept, X-API-Key, X-Requested-With")

		// Разрешаем кеширование preflight запросов (OPTIONS)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Разрешаем передавать куки и заголовки авторизации
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Если это OPTIONS запрос (preflight), сразу отвечаем
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Пропускаем запрос дальше
		c.Next()
	}
}
