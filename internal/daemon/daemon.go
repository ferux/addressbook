package daemon

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/ferux/addressbook/internal/controllers"
	"github.com/ferux/addressbook/internal/models"
	"github.com/gorilla/mux"
)

//MAXFILESIZE limits the maximum size of CSV file (used in import)
const MAXFILESIZE = 1024 * 1024 * 8

var controller controllers.User
var logger *log.Logger

//errorMessage struct used to return error messages back to client
type errorMessage struct {
	Err string `json:"error,omitempty"`
}

//Precompiled checks for email and names
var (
	emailCheck = regexp.MustCompile("^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*$").MatchString
	nameCheck  = regexp.MustCompile("^[a-zA-Z]+$").MatchString
)

//Start runs service. It accepts pointer to mgo.Collection for creating controller,
//INPUT:
// db (*mgo.Collection) -- points to the collection in database
// addr (string) 		-- defines listen address for income connections
// w (io.Writer)		-- Writer for logging.
func Start(db *mgo.Database, addr string, w io.Writer) error {
	logger = log.New(w, "[Daemon] ", log.Lshortfile+log.Ldate+log.Ltime)
	if db == nil {
		return errors.New("DB variable is nil")
	}
	if addr == "" {
		return errors.New("Address is empty")
	}
	if w == nil {
		w = ioutil.Discard
	}
	controller = controllers.User{Collection: db.C("users")}

	router := addRoutes()

	logger.Printf("Ready to accept connections on %s", addr)
	return http.ListenAndServe(addr, router)
}

func addRoutes() *mux.Router {
	router := mux.NewRouter()
	router.Use(sessionMiddleware)
	router.Use(logMiddleware)

	router.HandleFunc("/", helloHandler).Methods("GET")
	routerV1 := router.PathPrefix("/api/v1/addressbook").Subrouter()
	routerV1.HandleFunc("/", listUsersHandler).Methods("GET")
	routerV1.HandleFunc("/user", createUserHandler).Methods("POST")
	routerV1.HandleFunc("/user/{id}", selectUserHandler).Methods("GET")
	routerV1.HandleFunc("/user/{id}", updateUserHandler).Methods("PUT")
	routerV1.HandleFunc("/user/{id}", deleteUserHandler).Methods("DELETE")
	routerV1.HandleFunc("/export", downloadCSVHandler).Methods("GET")
	routerV1.HandleFunc("/import", uploadCSVHandler).Methods("PUT")
	routerV1.HandleFunc("/clear", clearHandler).Methods("GET")
	return router
}

type daemonKeys uint8

const (
	keySID daemonKeys = iota
)

type middlewareFunc func(w http.ResponseWriter, r *http.Request)

func (f middlewareFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

func sessionMiddleware(f http.Handler) http.Handler {
	m := func(w http.ResponseWriter, r *http.Request) {
		sidcookie, err := r.Cookie("sessionid")
		if err != nil {
			sidcookie = &http.Cookie{}
			logger.Printf("err getting cookie: %v", err)
			sid := rand.Uint64()
			logger.Printf("new sid: %d", sid)
			sidcookie.Value = strconv.FormatUint(sid, 10)
			sidcookie.Expires = time.Now().Add(time.Hour * 24 * 7)
			sidcookie.HttpOnly = true
			sidcookie.Name = "sessionid"
			sidcookie.Path = "/"
			http.SetCookie(w, sidcookie)
			ctx := context.WithValue(r.Context(), keySID, sid)
			r.WithContext(ctx)
		} else {
			sidcookie.Expires = time.Now().Add(time.Hour * 24 * 7)
		}
		f.ServeHTTP(w, r)
	}
	return middlewareFunc(m)
}

func logMiddleware(f http.Handler) http.Handler {
	m := func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("Request %s from %s", r.RequestURI, r.RemoteAddr)
		f.ServeHTTP(w, r)
	}
	return middlewareFunc(m)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Add("Content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	users, err := controller.ListUsers()
	if err != nil {
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Got an error trying to retrieve userlist. Reason: %v", err)
		return
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&users)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, jsonError(err), http.StatusBadRequest)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Can't parse request body. Reason: %v", err)
		return
	}
	if errs := checkCorrectValues(&user); errs != nil {
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errs)
		logger.Printf("Can't parse structure: %v", user)
		return
	}
	id, err := controller.CreateUser(&user)
	if err != nil {
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Can't create user. Reason: %v", err)
		return
	}
	logger.Printf("User has been created. ID: %s", id.String())
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-type", "application/json")
	json.NewEncoder(w).Encode(&models.User{ID: id})
}

func selectUserHandler(w http.ResponseWriter, r *http.Request) {
	varsID := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(varsID) {
		http.Error(w, jsonError(fmt.Errorf("Object %s is not ObjectID", varsID)), http.StatusBadRequest)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Object %s is not objectID", varsID)
		return
	}
	id := bson.ObjectIdHex(varsID)

	user, err := controller.SelectUser(id)
	if err != nil {
		w.Header().Add("Content-type", "application/json")
		status := http.StatusInternalServerError
		if err == mgo.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, jsonError(err), status)
		logger.Printf("Got an error while selecting user: %v", err)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	varsID := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(varsID) {
		w.Header().Add("Content-type", "application/json")
		http.Error(w, jsonError(fmt.Errorf("Object %s is not ObjectID", varsID)), http.StatusBadRequest)
		logger.Printf("Object %s is not objectID", varsID)
		return
	}
	user := models.User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, jsonError(err), http.StatusBadRequest)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Can't parse request body. Reason: %v", err)
		return
	}
	if errs := checkCorrectValues(&user); errs != nil {
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errs)
		logger.Printf("Can't parse structure: %v", user)
		return
	}
	id := bson.ObjectIdHex(varsID)
	user.ID = id

	if err := controller.UpdateUser(&user); err != nil {
		w.Header().Add("Content-type", "application/json")
		status := http.StatusInternalServerError
		if err == mgo.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, jsonError(err), status)
		logger.Printf("Can't update data. Reason: %v", len(id))
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	varsID := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(varsID) {
		w.Header().Add("Content-type", "application/json")
		http.Error(w, jsonError(fmt.Errorf("Object %s is not ObjectID", varsID)), http.StatusBadRequest)
		logger.Printf("Object %s is not objectID", varsID)
		return
	}
	id := bson.ObjectIdHex(varsID)

	err := controller.DeleteUser(id)
	if err != nil {
		w.Header().Add("Content-type", "application/json")
		status := http.StatusInternalServerError
		if err == mgo.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, jsonError(err), status)
		logger.Printf("Can't delete data. Reason: %v", len(id))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func uploadCSVHandler(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(r.Header.Get("Content-type")) != "text/csv" {
		http.Error(w, jsonError(errors.New("Content type should be text/csv")), http.StatusBadRequest)
		logger.Printf("Content type should be text/csv")
		return
	}
	records, err := csv.NewReader(io.LimitReader(r.Body, MAXFILESIZE)).ReadAll()
	if err != nil {
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Can't read file. Reason: %v", err)
		return
	}
	users, err := populateUsers(records)
	if err != nil {
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Can't read file.\nReason: %v", err)
		return
	}
	logger.Println("CSV file has been parsed. Uploading to database")
	switch r.Header.Get("Append-type") {
	case "clear":
		if err := controller.CleanRecords(); err != nil {
			logger.Printf("Can't clean records. Reason: %v", err)
			http.Error(w, jsonError(err), http.StatusInternalServerError)
			w.Header().Add("Content-type", "application/json")
			return
		}
		fallthrough
	case "append":
		for _, item := range users {
			if err := controller.UploadUser(&item); err != nil {
				http.Error(w, jsonError(err), http.StatusInternalServerError)
				w.Header().Add("Content-type", "application/json")
				logger.Printf("Can't upload user. Reason: %v", err)
				return
			}
		}
		break
	default:
		for _, item := range users {
			if err := controller.UpsertUser(&item); err != nil {
				http.Error(w, jsonError(err), http.StatusInternalServerError)
				w.Header().Add("Content-type", "application/json")
				logger.Printf("Can't upsert user. Reason: %v", err)
				return
			}
		}
		break
	}
	http.Redirect(w, r, "/api/v1/addressbook/", http.StatusFound)
}

func downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	users, err := controller.ListUsers()
	if err != nil {
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Got an error trying to retrieve userlist: %v", err)
		return
	}
	if len(*users) == 0 {
		http.Error(w, jsonError(errors.New("There is nothing to show")), http.StatusNoContent)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("There is nothing to show")
		return
	}

	records := [][]string{}
	for _, item := range *users {
		records = append(records, []string{item.ID.Hex(), item.FirstName, item.LastName, item.Email, item.Phone})
	}
	w.Header().Add("Content-type", "text/csv")
	w.Header().Add("Content-disposition", "attachment; filename=import.csv")
	w.WriteHeader(http.StatusOK)
	err = csv.NewWriter(w).WriteAll(records)
	if err != nil {
		logger.Printf("Got an error while writing records: %v", err)
	}
	logger.Println("CSV has been uploaded")
}

func clearHandler(w http.ResponseWriter, r *http.Request) {
	if err := controller.CleanRecords(); err != nil {
		http.Error(w, jsonError(err), http.StatusInternalServerError)
		w.Header().Add("Content-type", "application/json")
		logger.Printf("Can't clean records. Reason: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func jsonError(err error) string {
	return fmt.Sprintf(`{"error": "%s"}`, err.Error())
}

func populateUsers(input [][]string) ([]models.User, error) {

	if len(input) == 0 {
		return nil, errors.New("Input is empty")
	}
	users := []models.User{}
	for _, item := range input {
		if len(item) != 5 {
			continue
		}
		user := models.User{
			ID:        bson.ObjectIdHex(item[0]),
			FirstName: item[1],
			LastName:  item[2],
			Email:     item[3],
			Phone:     item[4],
		}
		if errs := checkCorrectValues(&user); errs != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func checkCorrectValues(u *models.User) *[]errorMessage {
	var errs []errorMessage
	if !emailCheck(u.Email) {
		errs = append(errs, errorMessage{Err: "Can't parse email"})
	}
	if !nameCheck(u.FirstName) {
		errs = append(errs, errorMessage{Err: "First Name is incorrect"})
	}
	if !nameCheck(u.LastName) {
		errs = append(errs, errorMessage{Err: "Last Name is incorrect"})
	}
	if len(errs) == 0 {
		return nil
	}
	return &errs
}
