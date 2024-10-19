package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/trisaptono/producer/usecase/user"
)

func ServerRoutes() {
	router := gin.Default()
	users := router.Group("/users/")
	{
		users.POST("/", user.CreateUser)
		users.GET("/", user.GetUsers)
		users.GET("/:id", user.GetUser)
		users.PUT("/:id", user.UpdateUser)
		users.DELETE("/:id", user.DeleteUser)
	}
	_ = router.Run()
}
