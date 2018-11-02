package controllers

import (
	"github.com/ferux/addressbook/internal/models"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const collection = "users"

// User controller type uses collection to manipulate data
type User struct{ DB *mgo.Database }

// CreateUser func
func (c *User) CreateUser(u *models.User) (bson.ObjectId, error) {
	return models.CreateUser(c.DB.C(collection), u)
}

// UpdateUser func
func (c *User) UpdateUser(u *models.User) error {
	return models.UpdateUser(c.DB.C(collection), u)
}

// DeleteUser func
func (c *User) DeleteUser(id bson.ObjectId) error {
	return models.DeleteUser(c.DB.C(collection), id)
}

// SelectUser func
func (c *User) SelectUser(id bson.ObjectId) (*models.User, error) {
	return models.SelectUser(c.DB.C(collection), id)
}

// ListUsers func
func (c *User) ListUsers() (*[]models.User, error) {
	return models.ListUsers(c.DB.C(collection))
}

// UploadUser func
func (c *User) UploadUser(u *models.User) error {
	return models.UploadUser(c.DB.C(collection), u)
}

// UpsertUser func
func (c *User) UpsertUser(u *models.User) error {
	return models.UpsertUser(c.DB.C(collection), u)
}

// CleanRecords func
func (c *User) CleanRecords() error {
	return models.CleanRecords(c.DB.C(collection))
}
