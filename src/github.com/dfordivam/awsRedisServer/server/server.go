package main

import (
	// Standard library packages
	"net/http"

	// Third party packages
	"github.com/dfordivam/awsRedisServer/controllers"
	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
)

func main() {
	// Instantiate a new router
	r := httprouter.New()

	userDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Get a UserController instance
	uc := controllers.NewUserController(userDB)

	// Get a user resource
	r.GET("/user/:id", uc.GetUser)

	r.POST("/user", uc.CreateUser)

	r.DELETE("/user/:id", uc.RemoveUser)

	// Fire up the server
	http.ListenAndServe("localhost:3000", r)
}
