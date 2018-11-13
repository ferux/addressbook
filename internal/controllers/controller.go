package controllers

import (
	"github.com/ferux/addressbook"
	"gopkg.in/mgo.v2"
)

// Controller stores connection to database.
type Controller struct {
	db     *mgo.Database
	status addressbook.Code
}

// NewController creates new instance of repo.
func NewController(db *mgo.Database) *Controller {
	return &Controller{db: db, status: addressbook.Running}
}

// User returns User collection.
func (c *Controller) User() *User {
	return &User{Collection: c.db.C(userCollection)}
}
