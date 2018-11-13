package controllers

import (
	"gopkg.in/mgo.v2"
)

// Controller stores connection to database.
type Controller struct {
	db *mgo.Database
}

// NewController creates new instance of repo.
func NewController(db *mgo.Database) *Controller {
	return &Controller{db: db}
}

// User returns User collection.
func (c *Controller) User() *User {
	return &User{Collection: c.db.C(userCollection)}
}
