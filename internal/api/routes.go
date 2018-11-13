package api

import (
	"encoding/json"
	"net/http"

	"github.com/ferux/addressbook"
	"github.com/gorilla/mux"
)

func (a *API) registerRoutes() *mux.Router {
	r := mux.NewRouter()

	r.Use(a.sessionControl, a.logRequests)

	r.HandleFunc("/status", handleServerStatus)

	r.NotFoundHandler = a.sessionControl(a.logRequests(a.notFoundHandler()))
	rv1 := r.PathPrefix("/api/v1/book").Subrouter()
	rv1.HandleFunc("", a.helloHandler).Methods("GET")
	rv1.HandleFunc("/", a.helloHandler).Methods("GET")
	rv1.HandleFunc("/user", a.listUsersHandler).Methods("GET")
	rv1.HandleFunc("/user", a.createUserHandler).Methods("POST")
	rv1.HandleFunc("/user/{id}", a.selectUserHandler).Methods("GET")
	rv1.HandleFunc("/user/{id}", a.updateUserHandler).Methods("PUT")
	rv1.HandleFunc("/user/{id}", a.deleteUserHandler).Methods("DELETE")
	rv1.HandleFunc("/export", a.downloadCSVHandler).Methods("GET")
	return r
}

func handleServerStatus(w http.ResponseWriter, _ *http.Request) {
	st := addressbook.MakeReport()
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&st)
}
