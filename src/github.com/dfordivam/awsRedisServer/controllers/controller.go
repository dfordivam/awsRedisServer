package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/dfordivam/awsRedisServer/models"
	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type (
	// UserController represents the controller for operating on the User resource
	UserController struct {
		userDB *redis.Client
	}
)

const UserNameHashKey = "UserNameHashKey"
const UserIDHashKey = "UserIDHashKey"
const UserCount = "UserCount"

func NewUserController(u *redis.Client) *UserController {
	if u.Exists(UserNameHashKey).Val() == 0 {
		auser := models.UserNameHashFieldValue{"adminpass", 1}
		au, _ := json.Marshal(auser)
		u.HSet(UserNameHashKey, "admin", au)
		fmt.Println("Created UserNameHashKey")
	}
	if u.Exists(UserIDHashKey).Val() == 0 {
		u.HSet(UserIDHashKey, "1", "admin")
		fmt.Println("Created UserIDHashKey")
	}
	if u.Exists(UserCount).Val() == 0 {
		u.Set(UserCount, 1, 0)
	}
	return &UserController{u}
}

// GetUser retrieves an individual user resource
func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Stub an example user
	u := models.User{
		Name: "Bob Smith",
		Pass: "passbob",
		Id:   1,
	}

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(u)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

// CreateUser creates a new user resource
func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Stub an user to be populated from the body
	u := models.User{}

	// Populate the user data
	json.NewDecoder(r.Body).Decode(&u)

	if uc.userDB.HExists(UserNameHashKey, u.Name).Val() {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(405)
		fmt.Fprintf(w, "%s", "Username not available")
		return
	}

	// Add an Id
	count := uc.userDB.Incr(UserCount)
	u.Id = count.Val()

	ufv, _ := json.Marshal(models.UserNameHashFieldValue{u.Name, u.Id})
	i := strconv.FormatInt(u.Id, 10)

	// Add to DBs
	uc.userDB.HSet(UserNameHashKey, u.Name, ufv)
	uc.userDB.HSet(UserIDHashKey, i, u.Name)

	// Marshal provided interface into JSON structure
	// uj, _ := json.Marshal(u)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", "Success")
}

// RemoveUser removes an existing user resource
// NYI
func (uc UserController) RemoveUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// TODO: only write status for now
	w.WriteHeader(200)
}
