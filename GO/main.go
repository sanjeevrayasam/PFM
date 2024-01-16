package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sanjeevrayasam/pfm-go/controllers"
	"github.com/sanjeevrayasam/pfm-go/docs"
	"github.com/sanjeevrayasam/pfm-go/models"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("../db.sqlite"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.User{})
	fmt.Println("Automigrations completed ...")
	models.DB = db

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	// Authentication routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/register", controllers.Register)
		v1.POST("/login", controllers.LoginUser)
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Run()
}
