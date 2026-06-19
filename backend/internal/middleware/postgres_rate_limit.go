package middleware

import (
	"net/http"
	"time"

	"backend/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
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

		// Buscar o registro por IP. Usar Find em vez de First para não registrar erro GORM "record not found"
		var attempt LoginAttempt
		config.DB.Where("ip = ?", ip).Limit(1).Find(&attempt)

		// Verificar se o registro já existe no banco
		exists := attempt.IP != ""

		if exists {
			// Se o reset_at já passou, o limite expirou. Resetar o contador localmente.
			if attempt.ResetAt.Before(now) {
				attempt.Count = 0
				attempt.ResetAt = now.Add(ttl)
			}
		} else {
			// Se não existe, inicializar o objeto local com contagem 0
			attempt = LoginAttempt{
				IP:      ip,
				Count:   0,
				ResetAt: now.Add(ttl),
			}
		}

		// Verificar se o limite de tentativas foi atingido
		if attempt.Count >= maxAttempts {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Tente novamente em 5 minutos.",
			})
			c.Abort()
			return
		}

		c.Next()

		status := c.Writer.Status()
		// Incrementar tentativas apenas se falhou (401 Unauthorized)
		if status == http.StatusUnauthorized {
			attempt.Count++

			// Salvar usando ON CONFLICT para evitar conflito de chave única se houver concorrência
			config.DB.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&attempt)
		} else if status == http.StatusOK || status == http.StatusCreated {
			// Se o login/registro foi bem sucedido, limpar o contador para esse IP no banco
			if exists && attempt.Count > 0 {
				config.DB.Model(&LoginAttempt{}).Where("ip = ?", ip).Update("count", 0)
			}
		}
	}
}
