package repo

import (
	"github.com/ferux/AddressBook/models"
	"gopkg.in/mgo.v2/bson"
)

// Users for interacting with User database
type Users interface {
	CreateUser(u *models.User) (bson.ObjectId, error)
	UpdateUser(u *models.User) error
	DeleteUser(id bson.ObjectId) error
	SelectUser(id bson.ObjectId) (*models.User, error)
	ListUsers() ([]*models.User, error)
}
