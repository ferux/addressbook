package apiV2

import (
	"time"

	"github.com/ferux/addressbook/internal/repo"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/rs/zerolog"
)

// APIV2 contains links to storages
type APIV2 struct {
	users repo.Users

	e      *echo.Echo
	logger zerolog.Logger
}

// NewAPIV2 inits new APIV2 instance.
func NewAPIV2(urepo repo.Users) *APIV2 {
	logger := zerolog.
		New(zerolog.NewConsoleWriter()).
		With().
		Str("pkg", "apiV2").
		Logger()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	return &APIV2{
		users:  urepo,
		e:      e,
		logger: logger,
	}
}

func (a *APIV2) sessionmw() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var sid string
			req := c.Request()

			sidcookie, err := c.Cookie("sessionid")
			if err != nil {
				sid = addSessionCookie(c)

			} else {
				sidcookie.Expires = time.Now().Add(time.Hour * 24)
				c.SetCookie(sidcookie)
				sid = sidcookie.Value
			}

			req = req.WithContext(WithSID(req.Context(), sid))
			c.SetRequest(req)

			return next(c)
		}
	}
}
