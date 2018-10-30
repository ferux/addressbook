package models

import (
	"log"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const testConnectionString = "mongodb://192.168.99.100:27017"

var collection *mgo.Collection

func InitConnection() {
	session, err := mgo.Dial(testConnectionString)
	if err != nil {
		log.Fatalf("Got unexpected error %v", err)

	}
	collection = session.DB("Test").C("Test")
}

func TestEverything(t *testing.T) {
	if collection == nil {
		InitConnection()
	}
	//User creation test
	t.Run("User creation test", testCreateUser)
	t.Run("User select test", testSelectUser)
	t.Run("User update test", testUpdateUser)
	t.Run("User delete test", testDeleteUser)

}

func testCreateUser(t *testing.T) {
	if collection == nil {
		InitConnection()
	}
	user1 := User{
		FirstName: "Test_FName",
		LastName:  "Test_LName",
		Email:     "Test_Email",
		Phone:     "Test_Phone",
	}

	_, err := CreateUser(nil, nil)
	if err == nil {
		t.Errorf("Should catch an error but got nothing")
		t.Fail()
	}
	_, err = CreateUser(collection, nil)
	if err == nil {
		t.Errorf("Should catch an error but got nothing")
		t.Fail()
	}
	_, err = CreateUser(collection, &user1)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
		t.Fail()
	}
}

func testSelectUser(t *testing.T) {
	if collection == nil {
		InitConnection()
	}
	t.Log("Check if data inserted correctly")
	var user1 User
	err := collection.Find(bson.M{}).One(&user1)
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
		t.Fail()
	}
	user2, err := SelectUser(collection, user1.ID)
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
		t.Fail()
	}
	if !reflect.DeepEqual(&user1, user2) {
		t.Error("Origin and inserted user should be equal")
		t.Errorf("Got: %v\nExpected: %v", user2, &user1)
		t.Fail()
	}
}

func testUpdateUser(t *testing.T) {
	if collection == nil {
		InitConnection()
	}
	user1 := User{
		FirstName: "Test_FName",
		LastName:  "Test_LName",
		Email:     "Test_Email",
		Phone:     "Test_Phone",
	}

	id, err := CreateUser(collection, &user1)
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
		t.Fail()
	}
	user1 = User{
		ID:        id,
		FirstName: "Test_FName1",
		LastName:  "Test_LName1",
		Email:     "Test_Email1",
		Phone:     "Test_Phone1",
	}

	t.Log("Check if data can be updated correctly")
	err = UpdateUser(collection, &user1)
	if err != nil {
		t.Errorf("Got unexpected error %v", err)
		t.Fail()
	}

	user2, _ := SelectUser(collection, user1.ID)
	if !reflect.DeepEqual(&user1, user2) {
		t.Error("Origin and inserted user should be equal")
		t.Errorf("Got: %v\nExpected: %v", user2, &user1)
		t.Fail()
	}
	CleanRecords(collection)
}

func testDeleteUser(t *testing.T) {
	if collection == nil {
		InitConnection()
	}

}

func testCleanRecords(t *testing.T) {
	CleanRecords(collection)
}
