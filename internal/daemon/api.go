package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ferux/addressbook/internal/controllers"
	"github.com/ferux/addressbook/internal/models"
	"github.com/ferux/addressbook/internal/types"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrIDInvalid reports in case user id is not a bson_id
	ErrIDInvalid = errors.New("not a valid id")
)

// API serves requests from clients.
// TODO: handle database reconnect
type API struct {
	db     *mgo.Database
	logger *logrus.Entry
	conf   types.API
}

// NewAPI creates new instance of API.
func NewAPI(dbconn *mgo.Database, apiconf types.API) *API {
	return &API{
		db:     dbconn,
		logger: logrus.New().WithField("pkg", "daemon"),
		conf:   apiconf,
	}
}

func (a *API) registerRoutes() *mux.Router {
	r := mux.NewRouter()

	r.Use(a.sessionmw, a.logmw)

	r.NotFoundHandler = a.sessionmw(a.logmw(a.notFoundHandler()))
	rv1 := r.PathPrefix("/api/v1/book").Subrouter()
	rv1.HandleFunc("", a.helloHandler).Methods("GET")
	rv1.HandleFunc("/", a.helloHandler).Methods("GET")
	rv1.HandleFunc("/user", a.listUsersHandler).Methods("GET")
	rv1.HandleFunc("/user/{id}", a.selectUserHandler).Methods("GET")
	rv1.HandleFunc("/user/{id}", a.updateUserHandler).Methods("PUT")
	rv1.HandleFunc("/user/{id}", a.deleteUserHandler).Methods("DELETE")

	return r
}

func addRoutes() *mux.Router {
	router := mux.NewRouter()
	routerV1 := router.PathPrefix("/api/v1/addressbook").Subrouter()
	routerV1.HandleFunc("/export", downloadCSVHandler).Methods("GET")
	routerV1.HandleFunc("/import", uploadCSVHandler).Methods("PUT")
	routerV1.HandleFunc("/clear", clearHandler).Methods("GET")
	return router
}

func (a *API) sessionmw(f http.Handler) http.Handler {
	m := func(w http.ResponseWriter, r *http.Request) {
		sidcookie, err := r.Cookie("sessionid")
		if err != nil {
			sidcookie = &http.Cookie{}
			a.logger.Debugf("err getting cookie: %v", err)
			sid := uuid.New().String()
			sidcookie.Value = sid
			sidcookie.Expires = time.Now().Add(time.Hour * 24 * 7)
			sidcookie.HttpOnly = true
			sidcookie.Name = "sessionid"
			sidcookie.Path = "/"
			http.SetCookie(w, sidcookie)
			r.WithContext(WithSID(r.Context(), sid))
		} else {
			sidcookie.Expires = time.Now().Add(time.Hour * 24 * 7)
		}
		f.ServeHTTP(w, r)
	}
	return middlewareFunc(m)
}

func (a *API) logmw(f http.Handler) http.Handler {
	m := func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		a.logger.WithFields(logrus.Fields{
			"request":   r.RequestURI,
			"address":   r.RemoteAddr,
			"method":    r.Method,
			"requestID": requestID,
		}).Info("accepted")
		ctx := WithRID(r.Context(), requestID)
		r.WithContext(ctx)
		f.ServeHTTP(w, r)
	}
	return middlewareFunc(m)
}

func (a *API) notFoundHandler() http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("invalid path %s", r.RequestURI)))
	}
	return middlewareFunc(h)
}

// Run runs API for serving
func (a *API) Run() error {
	a.logger.Info("starting api")
	router := a.registerRoutes()

	return http.ListenAndServe(a.conf.Listen, router)
}

func (a *API) helloHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "helloHandler",
	})
	logger.Info()
	w.Header().Add("content-type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (a *API) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "listUsersHandler",
	})
	logger.Info()
	users, err := (&controllers.User{DB: a.db}).ListUsers()
	if err != nil {
		http.Error(w, "can't get users list", http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		logger.WithError(err).Error("can't get users list")
		return
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&users)
}

func (a *API) createUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "createUserHandler",
	})
	logger.Info()
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, jsonError(err), http.StatusBadRequest)
		a.logger.WithError(err).Error("can't parse request")
		return
	}
	if errs := checkCorrectValues(&user); errs != nil {
		w.Header().Add("content-type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errs)
		logger.Error("errs", errs)
		return
	}
	id, err := (&controllers.User{DB: a.db}).CreateUser(&user)
	if err != nil {
		http.Error(w, "can't create user", http.StatusBadRequest)
		a.logger.WithError(err).Error("can't insert user")
		return
	}
	a.logger.WithField("userid", id).Info("user has been created")
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(&models.User{ID: id})
}

func findUser(r *http.Request) (user *models.User, err error) {
	varsID := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(varsID) {
		err := ErrIDInvalid
		return user, err
	}

	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		return user, err
	}
	if msgs := checkCorrectValues(user); msgs != nil {
		err = errors.New(strings.Join(msgs, ";"))
		return user, err
	}
	id := bson.ObjectIdHex(varsID)
	user.ID = id
	return user, err
}

func (a *API) selectUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "selectUserHandler",
	})
	logger.Info()
	user, err := findUser(r)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusBadRequest)
		logger.WithError(err).Error("can't get user")
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *API) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "updateUserHandler",
	})
	logger.Info()
	user, err := findUser(r)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusBadRequest)
		return
	}

	if err := controller.UpdateUser(user); err != nil {
		status := http.StatusInternalServerError
		if err == mgo.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, "something went wrong", status)
		logger.WithError(err).Error("can't update data")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (a *API) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "updateUserHandler",
	})
	logger.Info()
	varsID := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(varsID) {
		http.Error(w, ErrIDInvalid.Error(), http.StatusBadRequest)
		logger.WithError(ErrIDInvalid).Error()
		logger.Printf("Object %s is not objectID", varsID)
		return
	}
	id := bson.ObjectIdHex(varsID)

	err := controller.DeleteUser(id)
	if err != nil {
		w.Header().Add("Content-type", "text/plain")
		status := http.StatusBadRequest
		if err == mgo.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, "can't delete user", status)
		logger.WithError(err).WithField("id", id).Error("can't delete user")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id.String()))
}
