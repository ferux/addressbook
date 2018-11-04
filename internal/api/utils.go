package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/ferux/addressbook/internal/models"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

//MAXFILESIZE limits the maximum size of CSV file (used in import)
const MAXFILESIZE = 1024 * 1024 * 8

//Precompiled checks for email and names
var (
	emailCheck = regexp.MustCompile("^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*$").MatchString
	nameCheck  = regexp.MustCompile("^[a-zA-Z]+$").MatchString
)

// ResponseError stores information about request ID and corresponding error.
type ResponseError struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
	Code      int    `json:"code"`
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("reqid=%s msg=%s code=%d", e.RequestID, e.Message, e.Code)
}

func wrapError(msg string, r *http.Request, code int) *ResponseError {
	rid := GetRID(r.Context())
	return &ResponseError{
		RequestID: rid,
		Message:   msg,
		Code:      code,
	}
}

func checkCorrectValues(u models.User) []string {
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

func findUser(r *http.Request) (user *models.User, err error) {
	varsID := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(varsID) {
		err := ErrIDInvalid
		return user, err
	}
	defer r.Body.Close()
	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		return user, err
	}
	if msgs := checkCorrectValues(*user); msgs != nil {
		err = errors.New(strings.Join(msgs, ";"))
		return user, err
	}
	id := bson.ObjectIdHex(varsID)
	user.ID = id
	return user, err
}

// daemonKeys for context
type daemonKeys uint8

// enums for DaemonKeys
const (
	keySID daemonKeys = iota
	keyReqID
)

// WithSID adds sid to ctx
func WithSID(ctx context.Context, sid string) context.Context {
	ctxn := context.WithValue(ctx, keySID, sid)
	return ctxn
}

// GetSID retrieves SID from ctx
func GetSID(ctx context.Context) string {
	sid, _ := ctx.Value(keySID).(string)
	return sid
}

// WithRID adds sid to ctx
func WithRID(ctx context.Context, rid string) context.Context {
	ctx = context.WithValue(ctx, keyReqID, rid)
	return ctx
}

// GetRID retrieves SID from ctx
func GetRID(ctx context.Context) string {
	rid, _ := ctx.Value(keyReqID).(string)
	return rid
}

type middlewareFunc func(w http.ResponseWriter, r *http.Request)

func (f middlewareFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}
