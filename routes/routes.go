package routes

import (
	"net/http"

	"github.com/AgungRilo/bookshelf-app/controllers"
	"github.com/AgungRilo/bookshelf-app/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	api := r.Group("/api")
	api.POST("/users/login", controllers.Login)

	categories := api.Group("/categories", middlewares.JWTAuth())
	categories.GET("", controllers.GetAllCategories)
	categories.POST("", controllers.CreateCategory)
	categories.GET("/:id", controllers.GetCategoryByID)
	categories.PUT("/:id", controllers.UpdateCategory)
	categories.DELETE("/:id", controllers.DeleteCategory)
	categories.GET("/:id/books", controllers.GetBooksByCategory)

	books := api.Group("/books", middlewares.JWTAuth())
	books.GET("", controllers.GetAllBooks)
	books.POST("", controllers.CreateBook)
	books.GET("/:id", controllers.GetBookByID)
	books.PUT("/:id", controllers.UpdateBook)
	books.DELETE("/:id", controllers.DeleteBook)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint tidak ditemukan"})
	})

	return r
}
