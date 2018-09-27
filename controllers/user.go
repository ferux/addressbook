package controllers

import (
	"github.com/ferux/addressbook/models"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User controller type uses collection to manipulate data
type User struct {
	Collection *mgo.Collection
}

// CreateUser func
func (c *User) CreateUser(u *models.User) (bson.ObjectId, error) {
	return models.CreateUser(c.Collection, u)
}

// UpdateUser func
func (c *User) UpdateUser(u *models.User) error {
	return models.UpdateUser(c.Collection, u)
}

// DeleteUser func
func (c *User) DeleteUser(id bson.ObjectId) error {
	return models.DeleteUser(c.Collection, id)
}

// SelectUser func
func (c *User) SelectUser(id bson.ObjectId) (*models.User, error) {
	return models.SelectUser(c.Collection, id)
}

// ListUsers func
func (c *User) ListUsers() (*[]models.User, error) {
	return models.ListUsers(c.Collection)
}

// UploadUser func
func (c *User) UploadUser(u *models.User) error {
	return models.UploadUser(c.Collection, u)
}

func (c *User) UpsertUser(u *models.User) error {
	return models.UpsertUser(c.Collection, u)
}

func (c *User) CleanRecords() error {
	return models.CleanRecords(c.Collection)
}
