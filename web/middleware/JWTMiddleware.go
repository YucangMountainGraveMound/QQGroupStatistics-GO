package middleware

import (
	"time"

	"dormon.net/qq/web/controller"

	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"dormon.net/qq/config"
)

// JWTMiddleware JWT中间件
func JWTMiddleware() *jwt.GinJWTMiddleware {
	return &jwt.GinJWTMiddleware{
		Realm:         "qq",
		Key:           []byte(config.Config().Secret),
		Timeout:       30 * 24 * time.Hour,
		MaxRefresh:    time.Hour,
		Authenticator: controller.Authenticator,
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"dormon": "",
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}

}
