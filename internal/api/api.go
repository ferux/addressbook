package api

import (
	"github.com/ferux/addressbook/internal/types"

	"gopkg.in/mgo.v2"
)

// API is a struct for serving HTTP requests.
type API struct {
	DB   *mgo.Database
	conf *types.API
}

// NewAPI creates new api.
func NewAPI(db *mgo.Database, conf *types.API) *API {
	return &API{
		DB:   db,
		conf: conf,
	}
}

// Run starts listening on selected port and serve new requests.
func (a *API) Run() error {
	return nil
}
