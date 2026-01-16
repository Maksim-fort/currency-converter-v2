package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck возвращает статус здоровья сервиса
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "currency-converter-api",
		"version": "v2.0.0",
	})
}
