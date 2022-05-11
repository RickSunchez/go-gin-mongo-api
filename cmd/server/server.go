package main

import (
	"flag"
	"fmt"
	"gin-server/internal/api"

	"github.com/gin-gonic/gin"
)

func main() {
	port := flag.String("p", "8080", "server port")

	flag.Parse()

	router := gin.Default()

	router.GET("/", api.MethodsList)
	router.POST("/create", api.CreateUser)
	router.POST("/make_friends", api.MakeFriends)
	router.DELETE("/user", api.DeleteUser)
	router.GET("/friends/:user_id", api.GetFriends)
	router.PUT("/:user_id", api.EditAge)

	router.Run(":" + *port)
	fmt.Println(*port)
}
