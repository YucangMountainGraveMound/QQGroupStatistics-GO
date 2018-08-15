package web

import (
	"dormon.net/qq/web/controller"
	mv "dormon.net/qq/web/middleware"

	"github.com/gin-gonic/gin"
	"net/http"
)

// InitialRouter initial a router
func InitialRouter() *gin.Engine {
	router := gin.Default()
	router.Use(mv.CORSMiddleware())

	// No auth
	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"dormon":  "🍻🍻🍻🍻🍻🍻",
				"message": "Hello World!",
			})
		})
		api.POST("/login", mv.JWTMiddleware().LoginHandler)
		api.GET("/histogram/heatmap", controller.Histogram)
		api.Static("/image/raw", "images")

		api.POST("/record/coolq", controller.CoolQ)

		api.GET("/robot/:param", controller.RobotMessageCount)
	}

	// Require auth
	api.Use(mv.JWTMiddleware().MiddlewareFunc())
	histogram := api.Group("/histogram")
	{
		histogram.GET("/active_day", controller.UsersByDay)
	}

	message := api.Group("/message")
	{
		message.GET("/user", controller.MessageCountByUser)
		message.GET("/specific_time", controller.MessageCountBySpecificTime)
		message.GET("/terms", controller.MessageByTerms)
		message.GET("/habit", controller.MessageHabit)
	}

	frequency := api.Group("/frequency")
	{
		frequency.GET("/word", controller.WordFrequency)
	}

	image := api.Group("/image")
	{

		image.GET("/count_user", controller.ImagesCountWithUser)
	}

	relation := api.Group("/relation")
	{
		relation.GET("/at", controller.UserAt)
	}

	settings := api.Group("/settings")
	{
		settings.PUT("/dict", controller.SetDictionary)
		settings.GET("/dict", controller.GetDictionary)
	}

	user := api.Group("/user")
	{
		user.GET("/me", controller.Me)
		user.PUT("/password", controller.UpdatePassword)
	}

	return router
}
