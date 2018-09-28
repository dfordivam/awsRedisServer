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

const userNameHashKey = "userNameHashKey"
const userIDHashKey = "userIDHashKey"
const userCount = "userCount"
const userSessionDB = "userSessionDB"
const messageList = "messageList"

func NewUserController(u, a, s *redis.Client) *UserController {
	if u.Exists(userNameHashKey).Val() == 0 {
		auser := models.UserNameHashFieldValue{"adminpass", 1}
		au, _ := json.Marshal(auser)
		u.HSet(userNameHashKey, "admin", au)
		fmt.Println("Created userNameHashKey")
	}
	if u.Exists(userIDHashKey).Val() == 0 {
		u.HSet(userIDHashKey, "1", "admin")
		fmt.Println("Created userIDHashKey")
	}
	if u.Exists(userCount).Val() == 0 {
		u.Set(userCount, 1, 0)
	}
	rand.Seed(time.Now().UnixNano())
	s.FlushDBAsync()
	return &UserController{u, a, s}
}

// GetUser retrieves an individual user resource
func (uc UserController) LoginUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	fmt.Println("Doing login")
	// Stub an user to be populated from the body
	u := models.User{}

	// Populate the user data
	json.NewDecoder(r.Body).Decode(&u)

	if uc.mainDB.HExists(userNameHashKey, u.Name).Val() == false {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(405)
		fmt.Fprintf(w, "%s", "Username not valid")
		return
	}

	ufvJson := uc.mainDB.HGet(userNameHashKey, u.Name).Val()

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

	if uc.mainDB.HExists(userNameHashKey, u.Name).Val() {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(405)
		fmt.Fprintf(w, "%s", "Username not available")
		return
	}

	// Add an Id
	count := uc.mainDB.Incr(userCount)
	u.Id = count.Val()

	ufv, _ := json.Marshal(models.UserNameHashFieldValue{u.Pass, u.Id})
	i := strconv.FormatInt(u.Id, 10)

	// Add to DBs
	uc.mainDB.HSet(userNameHashKey, u.Name, ufv)
	uc.mainDB.HSet(userIDHashKey, i, u.Name)

	// Marshal provided interface into JSON structure
	// uj, _ := json.Marshal(u)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", "Success")
}

func (uc UserController) LogoutUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	fmt.Println("Doing logout")
	authHead, found := r.Header["Authorization"]
	sessTok := strings.TrimPrefix(authHead[0], "Bearer ")
	if found == false || sessTok == authHead[0] {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(401)
		fmt.Fprintf(w, "%s", "Not logged in")
		return
	}

	_, res := uc.sessionDB.Get(sessTok).Result()
	if res == redis.Nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(401)
		fmt.Fprintf(w, "%s", "Invalid auth token, login again")
		return
	}

	uc.sessionDB.Del(sessTok)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", "Success")
}

func (uc UserController) doAuthGetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) (user string) {

	authHead, found := r.Header["Authorization"]
	sessTok := strings.TrimPrefix(authHead[0], "Bearer ")
	if found == false || sessTok == authHead[0] {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(401)
		fmt.Fprintf(w, "%s", "Not logged in")
		return
	}

	userIdStr, _ := uc.sessionDB.Get(sessTok).Result()
	// TODO ->
	// if res == redis.Nil {
	// 	w.Header().Set("Content-Type", "text/plain")
	// 	w.WriteHeader(401)
	// 	fmt.Fprintf(w, "%s", "Invalid auth token, login again")
	// 	return
	// }

	user = uc.mainDB.HGet(userIDHashKey, userIdStr).Val()
	return
}

func (uc UserController) PostMessage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	user := uc.doAuthGetUser(w, r, p)

	var msgObj models.MessageObject
	var postMsg models.PostMessage
	json.NewDecoder(r.Body).Decode(&postMsg)

	msgObj.User = user
	msgObj.Message = postMsg.Message

	fmt.Println("Adding msg to db", msgObj)
	msg, _ := json.Marshal(msgObj)
	uc.mainDB.LPush(messageList, msg)

	uc.sendMessages(w, r, p, postMsg.LastSyncVal)
}

func (uc UserController) GetMessages(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	uc.doAuthGetUser(w, r, p)

	// The length of messageList when last synced
	// It acts as a pseudoId
	lastStr := p.ByName("last")

	var last int64
	if lastStr != "" {
		last, _ = strconv.ParseInt(lastStr, 10, 64)
	} else {
		last = 0
	}

	fmt.Println("last:-", lastStr, last)
	uc.sendMessages(w, r, p, last)
}

func (uc UserController) sendMessages(w http.ResponseWriter, r *http.Request, p httprouter.Params, last int64) {
	// []string

	var l int64
	length := uc.mainDB.LLen(messageList)
	if last != 0 {
		l = length.Val() - last - 1
	} else {
		// if last sync not available, then send a max of 50 messages
		l = 50
	}

	var sm models.SendMessages
	if l >= 0 {
		ss := uc.mainDB.LRange(messageList, 0, l).Val()

		ll := len(ss)
		msgs := make([]models.MessageObject, ll)
		for i := 0; i < ll; i++ {
			json.NewDecoder(strings.NewReader(ss[i])).Decode(&(msgs[ll-i-1]))
		}
		sm.Messages = msgs
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	// Data synced till this length
	sm.MessageId = length.Val()
	msgsJson, _ := json.Marshal(sm)
	fmt.Fprintf(w, "%s", msgsJson)
}

// Util stuff
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
