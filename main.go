package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// 학식 API
	meal := r.Group("meal")
	{
		meal.GET("/ids", GetMealIds)
	}
	r.Run()
}
