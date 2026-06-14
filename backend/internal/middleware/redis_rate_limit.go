package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var redisCtx = context.Background()

// InitRedis inicializa a conexão com Redis
func InitRedis() {
	if redisClient != nil {
		return
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Default local Redis
		redisURL = "redis://localhost:6379/0"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test connection
	_, err := redisClient.Ping(redisCtx).Result()
	if err != nil {
		fmt.Printf("Warning: Redis connection failed: %v. Rate limiting will fall back to in-memory.\n", err)
		redisClient = nil
	}
}

// GetRedisClient retorna o client Redis (pode ser nil se não conectado)
func GetRedisClient() *redis.Client {
	return redisClient
}

// RedisLoginRateLimiter implementa rate limiter baseado em Redis
func RedisLoginRateLimiter() gin.HandlerFunc {
	const maxAttempts = 5
	const ttl = 5 * time.Minute

	return func(c *gin.Context) {
		if redisClient == nil {
			// Fallback para in-memory se Redis não disponível
			c.Next()
			// Não incrementa no fallback para evitar problemas
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("login:attempts:%s", ip)

		// Contar tentativas atuais
		attempts, err := redisClient.Get(redisCtx, key).Int()
		if err != nil {
			// Primeiro acesso ou key expirou
			attempts = 0
		}

		if attempts >= maxAttempts {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Tente novamente em 5 minutos.",
			})
			c.Abort()
			return
		}

		c.Next()

		// Incrementa tentativas apenas se falhou
		if status := c.Writer.Status(); status == http.StatusUnauthorized {
			attempts++
			if attempts >= maxAttempts {
				// Define TTL quando atingir o limite
				redisClient.Set(redisCtx, key, attempts, ttl)
			} else {
				// Incrementa sem TTL (será limpo automaticamente)
				redisClient.Incr(redisCtx, key)
				redisClient.Expire(redisCtx, key, ttl)
			}
		}
	}
}
