package middleware

import (
	"net/http"
	"time"

	"backend/config"

	"github.com/gin-gonic/gin"
)

// LoginAttempt representa uma tentativa de login no banco de dados
type LoginAttempt struct {
	IP       string    `gorm:"primaryKey"`
	Count    int       `gorm:"default:0"`
	ResetAt  time.Time `gorm:"index"`
}

// InitPostgresRateLimiter inicializa a tabela de rate limiting no PostgreSQL
func InitPostgresRateLimiter() {
	// AutoMigrate para garantir que a tabela existe
	config.DB.AutoMigrate(&LoginAttempt{})
}

// PostgresLoginRateLimiter implementa rate limiter baseado em PostgreSQL
func PostgresLoginRateLimiter() gin.HandlerFunc {
	const maxAttempts = 5
	const ttl = 5 * time.Minute

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		// Verificar se existe uma tentativa ativa e se o TTL expirou
		var attempt LoginAttempt
		result := config.DB.Where("ip = ? AND reset_at > ?", ip, now).First(&attempt)

		if result.Error != nil {
			// TTL expirou ou não existe - resetar
			attempt = LoginAttempt{
				IP:      ip,
				Count:   0,
				ResetAt: now.Add(ttl),
			}
			config.DB.Create(&attempt)
		}

		// Verificar limite
		if attempt.Count >= maxAttempts {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Tente novamente em 5 minutos.",
			})
			c.Abort()
			return
		}

		c.Next()

		// Incrementar tentativas apenas se falhou (401 Unauthorized)
		if status := c.Writer.Status(); status == http.StatusUnauthorized {
			attempt.Count++

			// Atualizar ou criar
			if result.Error != nil {
				// Não existia antes
				config.DB.Create(&attempt)
			} else {
				// Atualizar existente
				config.DB.Save(&attempt)
			}
		}
	}
}
