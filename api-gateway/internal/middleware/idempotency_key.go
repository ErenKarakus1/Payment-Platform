package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireIdempotencyKey() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idempotencyKey := ctx.GetHeader("Idempotency-Key")
		idempotencyKey = strings.TrimSpace(idempotencyKey)
		if idempotencyKey == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "idempotency key is required"})
			return
		}
		parsedIdempotencyKey, err := uuid.Parse(idempotencyKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid idempotency key"})
			return
		}
		ctx.Request.Header.Set("Idempotency-Key", parsedIdempotencyKey.String())
		ctx.Next()
	}
}
