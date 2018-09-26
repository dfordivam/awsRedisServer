package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/dfordivam/awsRedisServer/models"
	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	// UserController represents the controller for operating on the User resource
	UserController struct {
		mainDB     *redis.Client
		activityDB *redis.Client
		sessionDB  *redis.Client
	}
)

const UserNameHashKey = "UserNameHashKey"
const UserIDHashKey = "UserIDHashKey"
const UserCount = "UserCount"
const UserSessionDB = "UserSessionDB"

func NewUserController(u, a, s *redis.Client) *UserController {
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
	rand.Seed(time.Now().UnixNano())
	s.FlushDBAsync()
	return &UserController{u, a, s}
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

// GetUser retrieves an individual user resource
func (uc UserController) LoginUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// Stub an user to be populated from the body
	u := models.User{}

	// Populate the user data
	json.NewDecoder(r.Body).Decode(&u)

	if uc.mainDB.HExists(UserNameHashKey, u.Name).Val() == false {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(405)
		fmt.Fprintf(w, "%s", "Username not valid")
		return
	}

	ufvJson := uc.mainDB.HGet(UserNameHashKey, u.Name).Val()

	var ufv models.UserNameHashFieldValue
	json.NewDecoder(strings.NewReader(ufvJson)).Decode(&ufv)

	if ufv.Pass != u.Pass {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(405)
		fmt.Fprintf(w, "%s", "Invalid credentials")
		return
	}

	// Create a random id as session token
	sessTok := randStringBytes(64)
	uc.sessionDB.Set(sessTok, ufv.Id, 24*time.Hour)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", sessTok)
}

// CreateUser creates a new user resource
func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Stub an user to be populated from the body
	u := models.User{}

	// Populate the user data
	json.NewDecoder(r.Body).Decode(&u)

	if uc.mainDB.HExists(UserNameHashKey, u.Name).Val() {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(405)
		fmt.Fprintf(w, "%s", "Username not available")
		return
	}

	// Add an Id
	count := uc.mainDB.Incr(UserCount)
	u.Id = count.Val()

	ufv, _ := json.Marshal(models.UserNameHashFieldValue{u.Name, u.Id})
	i := strconv.FormatInt(u.Id, 10)

	// Add to DBs
	uc.mainDB.HSet(UserNameHashKey, u.Name, ufv)
	uc.mainDB.HSet(UserIDHashKey, i, u.Name)

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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
