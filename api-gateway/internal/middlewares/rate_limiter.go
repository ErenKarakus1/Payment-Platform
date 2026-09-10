package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(client *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		merchantID := ctx.GetHeader("X-Merchant-ID")
		if merchantID == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "merchant id is required"})
			return
		}
		parsedMerchantID, err := uuid.Parse(merchantID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
			return
		}
		key := "rate_limit:merchant:" + parsedMerchantID.String() + ":" + ctx.FullPath()
		count, err := client.Incr(ctx.Request.Context(), key).Result()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if count == 1 {
			if err := client.Expire(ctx.Request.Context(), key, window).Err(); err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}

		if count > int64(limit) {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		ctx.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(limit-int(count)))
		ctx.Next()
	}
}
