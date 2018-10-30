package models

import (
	"errors"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//User struct describes the structure of user object
type User struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	FirstName string        `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName  string        `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Email     string        `json:"email,omitempty" bson:"email,omitempty"`
	Phone     string        `json:"phone,omitempty" bson:"phone,omitempty"`
}

//CreateUser creates a new user and put it to the database
func CreateUser(db *mgo.Collection, u *User) (bson.ObjectId, error) {
	if u == nil {
		return "", errors.New("Nil pointer to User struct")
	}
	u.ID = bson.NewObjectId()
	if err := db.Insert(&u); err != nil {
		return "", err
	}
	return u.ID, nil
}

//UploadUser creates a new user and put it to the database
func UploadUser(db *mgo.Collection, u *User) error {
	if u == nil {
		return errors.New("Nil pointer to User struct")
	}
	return db.Insert(&u)
}

//UpsertUser inserts or updates user record if the item with the same ID is exists
func UpsertUser(db *mgo.Collection, u *User) error {
	if u == nil {
		return errors.New("Nil pointer to User struct")
	}
	_, err := db.UpsertId(&u.ID, &u)
	return err
}

//SelectUser returns a user with specified id
func SelectUser(db *mgo.Collection, id bson.ObjectId) (*User, error) {
	u := User{}
	if err := db.FindId(id).One(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

//UpdateUser updates a user info
func UpdateUser(db *mgo.Collection, u *User) error {
	return db.UpdateId(u.ID, &u)
}

//DeleteUser removes user with specified id
func DeleteUser(db *mgo.Collection, id bson.ObjectId) error {
	return db.RemoveId(id)
}

//ListUsers returns the list of all users
func ListUsers(db *mgo.Collection) (*[]User, error) {
	users := make([]User, 0, 1)
	if err := db.Find(nil).All(&users); err != nil {
		return nil, err
	}
	return &users, nil
}

//CleanRecords erases all records at the collection
func CleanRecords(db *mgo.Collection) error {
	_, err := db.RemoveAll(bson.M{})
	return err
}
