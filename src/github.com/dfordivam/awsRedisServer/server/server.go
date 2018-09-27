package main

import (
	// Standard library packages
	"net/http"

	// Third party packages
	"github.com/dfordivam/awsRedisServer/controllers"
	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

func main() {
	// Instantiate a new router
	r := httprouter.New()

	mainDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	activityDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	sessionDB := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       2,  // use default DB
	})
	// Get a UserController instance
	uc := controllers.NewUserController(mainDB, activityDB, sessionDB)

	// User stuff
	r.POST("/auth/login", uc.LoginUser)

	r.POST("/auth/logout", uc.LogoutUser)

	// Messaging
	r.POST("/message", uc.PostMessage)

	r.GET("/messages/:last", uc.GetMessages)

	handler := cors.AllowAll().Handler(r)
	// Fire up the server
	http.ListenAndServe("localhost:3000", handler)
}
