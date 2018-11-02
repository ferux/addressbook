package api

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func addRoutes() *mux.Router {
	router := mux.NewRouter()
	router.Use(sessionMiddleware)
	router.Use(logMiddleware)

	// router.HandleFunc("/", helloHandler).Methods("GET")
	// routerV1 := router.PathPrefix("/api/v1/addressbook").Subrouter()
	// routerV1.HandleFunc("/", listUsersHandler).Methods("GET")
	// routerV1.HandleFunc("/user", createUserHandler).Methods("POST")
	// routerV1.HandleFunc("/user/{id}", selectUserHandler).Methods("GET")
	// routerV1.HandleFunc("/user/{id}", updateUserHandler).Methods("PUT")
	// routerV1.HandleFunc("/user/{id}", deleteUserHandler).Methods("DELETE")
	// routerV1.HandleFunc("/export", downloadCSVHandler).Methods("GET")
	// routerV1.HandleFunc("/import", uploadCSVHandler).Methods("PUT")
	// routerV1.HandleFunc("/clear", clearHandler).Methods("GET")
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
