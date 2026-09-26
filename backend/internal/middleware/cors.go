package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AturCORS mengatur header Cross-Origin Resource Sharing sesuai domain frontend yang diizinkan.
func AturCORS(urlFrontend string) gin.HandlerFunc {
	return func(c *gin.Context) {
		asalPermintaan := c.Request.Header.Get("Origin")

		// Jika asal permintaan cocok dengan URL frontend, domain vercel, atau wildcard
		if asalPermintaan != "" && (asalPermintaan == urlFrontend || urlFrontend == "*" || strings.HasSuffix(asalPermintaan, ".vercel.app") || asalPermintaan == "https://sivilize-demo-traffic.vercel.app") {
			c.Writer.Header().Set("Access-Control-Allow-Origin", asalPermintaan)
		} else if asalPermintaan != "" && (asalPermintaan == "http://localhost:3000" || asalPermintaan == "http://127.0.0.1:3000") {
			c.Writer.Header().Set("Access-Control-Allow-Origin", asalPermintaan)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
