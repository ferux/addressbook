package api

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
// TODO: copy session before getting data from db id:22 gh:15
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

func (a *API) sessionmw(f http.Handler) http.Handler {
	m := func(w http.ResponseWriter, r *http.Request) {
		sidcookie, err := r.Cookie("sessionid")
		if err != nil {
			sid := addSessionCookie(w)
			r = r.WithContext(WithSID(r.Context(), sid))
		} else {
			sidcookie.Expires = time.Now().Add(time.Hour * 24 * 7)
			http.SetCookie(w, sidcookie)
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
		r = r.WithContext(ctx)
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
	a.logger.WithField("listen", a.conf.Listen).Info("starting api")
	router := a.registerRoutes()
	return http.ListenAndServe(a.conf.Listen, router)
}

func (a *API) handleError(err error, w http.ResponseWriter) {
	switch e := err.(type) {
	case nil:
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		w.Write([]byte("unknown error"))
	case *ResponseError:
		if e.GetOrigin() == mgo.ErrNotFound || e.GetOrigin() == models.ErrAlreadyExists {
			a.handleError(e.GetOrigin(), w)
			return
		}
		w.WriteHeader(e.Code)
		w.Header().Add("content-type", "application/json")
		if errJSON := json.NewEncoder(w).Encode(e); errJSON != nil {
			a.logger.WithError(errJSON).Info("can't encode")
		}
	default:
		switch err {
		case models.ErrAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		case mgo.ErrNotFound:
			w.WriteHeader(http.StatusNotFound)
			w.Header().Add("content-type", "text/plain")
			w.Write([]byte("not found"))
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Add("content-type", "text/plain")
			w.Write([]byte(err.Error()))
		}

	}
}

func (a *API) helloHandler(w http.ResponseWriter, r *http.Request) {
	a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "helloHandler",
	}).Info()
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
		logger.WithError(err).Error("can't get users list")
		err = wrapError("error getting userlist", r, http.StatusInternalServerError, err)
		a.handleError(err, w)
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
		a.logger.WithError(err).Error("can't parse request")
		err = wrapError("error parsing body", r, http.StatusBadRequest, err)
		a.handleError(err, w)
		return
	}
	// TODO: handle this errors better. id:24 gh:16
	if errs := checkCorrectValues(user); errs != nil {
		logger.Error("errs", errs)
		w.Header().Add("content-type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errs)
		return
	}
	id, err := (&controllers.User{DB: a.db}).CreateUser(&user)
	if err != nil {
		a.logger.WithError(err).Error("can't insert user")
		err = wrapError("can't create user", r, http.StatusInternalServerError, err)
		a.handleError(err, w)
		return
	}
	user.ID = id
	a.logger.WithField("userid", id).Info("user has been created")
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(&user)
}

func (a *API) selectUserHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "selectUserHandler",
	})
	logger.Info()
	varsID := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(varsID) {
		logger.WithError(ErrIDInvalid).Error("invalid OID")
		err := wrapError(ErrIDInvalid.Error(), r, http.StatusBadRequest, nil)
		a.handleError(err, w)
		return
	}

	c := &controllers.User{DB: a.db}
	id := bson.ObjectIdHex(varsID)
	user, err := c.SelectUser(id)
	if err != nil {
		logger.WithError(err).Error("can't get user")
		err = wrapError("can't get user", r, http.StatusBadRequest, err)
		a.handleError(err, w)
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
	// TODO: better error handling. id:29 gh:18
	if err != nil {
		http.Error(w, "something went wrong", http.StatusBadRequest)
		return
	}

	if err := (&controllers.User{DB: a.db}).UpdateUser(user); err != nil {
		logger.WithError(err).Error("can't update data")
		err = wrapError("can't update user's info", r, http.StatusBadRequest, err)
		a.handleError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "applicaton/json")
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
		logger.WithError(ErrIDInvalid).Error("invalid OID")
		err := wrapError(ErrIDInvalid.Error(), r, http.StatusBadRequest, nil)
		a.handleError(err, w)
		return
	}

	id := bson.ObjectIdHex(varsID)
	err := (&controllers.User{DB: a.db}).DeleteUser(id)
	if err != nil {
		logger.WithError(err).Error("can't delete user")
		err = wrapError("can't delete user", r, http.StatusBadRequest, err)
		a.handleError(err, w)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(&models.User{ID: id})
}

// todo: rework
func (a *API) downloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.WithFields(logrus.Fields{
		"requestID": GetRID(r.Context()),
		"fn":        "updateUserHandler",
	})
	users, err := (&controllers.User{DB: a.db}).ListUsers()
	if err != nil {
		logger.WithError(err).Error("can't get users list")
		err = wrapError("error getting userlist", r, http.StatusInternalServerError, err)
		a.handleError(err, w)
		return
	}

	records := [][]string{}
	for _, item := range users {
		records = append(records, []string{item.ID.Hex(), item.FirstName, item.LastName, item.Email, item.Phone})
	}
	w.Header().Add("Content-type", "text/csv")
	w.Header().Add("Content-disposition", "attachment; filename=import.csv")
	w.WriteHeader(http.StatusOK)
	err = csv.NewWriter(w).WriteAll(records)
	if err != nil {
		logger.WithError(err).Error("can't write records")
		return
	}
}
