package apiV2

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
)

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

func addSessionCookie(c echo.Context) string {
	cookie := &http.Cookie{}
	sid := uuid.New().String()
	cookie.Value = sid
	cookie.Expires = time.Now().Add(time.Hour * 24 * 7)
	cookie.HttpOnly = true
	cookie.Name = "sessionid"
	cookie.Path = "/"
	c.SetCookie(cookie)
	return sid
}
