package daemon

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/ferux/addressbook/internal/controllers"
	"github.com/ferux/addressbook/internal/models"
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
	logger = log.New(w, "Daemon ", log.Lshortfile+log.Ldate+log.Ltime)
	if db == nil {
		return errors.New("DB variable is nil")
	}
	if addr == "" {
		return errors.New("Address is empty")
	}
	if w == nil {
		w = ioutil.Discard
	}
	controller = controllers.User{DB: db}

	router := addRoutes()

	logger.Printf("Ready to accept connections on %s", addr)
	return http.ListenAndServe(addr, router)
}

type daemonKeys uint8

const (
	keySID daemonKeys = iota
)

type middlewareFunc func(w http.ResponseWriter, r *http.Request)

func (f middlewareFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
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

func checkCorrectValues(u *models.User) []string {
	msgs := make([]string, 0, 3)
	if !emailCheck(u.Email) {
		msgs = append(msgs, "Email is incorrect")
	}
	if !nameCheck(u.FirstName) {
		msgs = append(msgs, "First Name is incorrect")
	}
	if !nameCheck(u.LastName) {
		msgs = append(msgs, "Last Name is incorrect")
	}
	if len(msgs) == 0 {
		return nil
	}
	return msgs
}
