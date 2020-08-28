package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/s-owl/skhuinfo-api/docs"
)

func main() {
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "SKHUINFO"
	docs.SwaggerInfo.Description = "SKHUINFO API"
	docs.SwaggerInfo.Version = "0.1"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r := gin.Default()
	api := r.Group("api")
	{
		v1 := api.Group("v1")
		// 학식 API
		meal := v1.Group("meal")
		{
			meal.GET("/ids", GetMealIds)
		}
		v1.GET("/schedules/:year/:month", GetSchedules)
	}

	// swagger handler
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run()
}
