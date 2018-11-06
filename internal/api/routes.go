package api

import "github.com/gorilla/mux"

func (a *API) registerRoutes() *mux.Router {
	r := mux.NewRouter()

	r.Use(a.sessionmw, a.logmw)

	r.NotFoundHandler = a.sessionmw(a.logmw(a.notFoundHandler()))
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
