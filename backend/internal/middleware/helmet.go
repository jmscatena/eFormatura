package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// HelmetMiddleware implementa security headers estilo Helmet.js
// Protege contra XSS, clickjacking, MIME sniffing, SSL stripping
func HelmetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. X-Frame-Options — previne clickjacking (embed em iframes)
		c.Header("X-Frame-Options", "DENY")

		// 2. X-Content-Type-Options — previne MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// 3. X-XSS-Protection — ativa XSS filter no browser
		c.Header("X-XSS-Protection", "1; mode=block")

		// 4. Strict-Transport-Security — força HTTPS por 2 anos
		c.Header("Strict-Transport-Security",
			"max-age=63072000; includeSubDomains; preload")

		// 5. Referrer-Policy — controla envio de referrer
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 6. Permissions-Policy — restringe acesso a features do browser
		c.Header("Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), payment=()")

		// 7. Content-Security-Policy — previne XSS e injection
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; "+
			"font-src 'self' data:; "+
			"connect-src 'self' https://*; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self' https://*;")

		// 8. Cache-Control para responses sensíveis
		if isSensitivePath(c.Request.URL.Path) {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// isSensitivePath identifica rotas que contêm dados sensíveis
func isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/login",
		"/register",
		"/api/",
		"/admin/",
		"/static/",
	}
	for _, p := range sensitivePaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
