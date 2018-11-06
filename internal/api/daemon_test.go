package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/ferux/AddressBook/models"
)

const testConnectionString = "mongodb://192.168.99.100:27017"

var userslist []models.User

// TODO: adapt to API.
//To run this tests the API server should be on.
func TestAll(t *testing.T) {
	userslist = make([]models.User, 0)
	defer t.Run("Test clear db", testClear)
	if !t.Run("Test clear db", testClear) {
		t.FailNow()
	}
	if !t.Run("Test create user", testCreateUser) {
		t.FailNow()
	}
	if !t.Run("Test list users", testListUsers) {
		t.FailNow()
	}
	if !t.Run("Test select user", testSelectUser) {
		t.FailNow()
	}
	if !t.Run("Test update user", testUpdateUser) {
		t.FailNow()
	}
	if !t.Run("Test remove user", testRemoveUser) {
		t.FailNow()
	}
	if !t.Run("Test list users", testListUsers) {
		t.FailNow()
	}
}

func findUserTest(id bson.ObjectId) models.User {
	for _, item := range userslist {
		if item.ID == id {
			return item
		}
	}
	return models.User{}
}

func removeUser(id bson.ObjectId) {
	for i, item := range userslist {
		if item.ID == id {
			userslist = append(userslist[:i], userslist[i+1:]...)
		}
	}
}

func clearUsersList() {
	userslist = make([]models.User, 0)
}

func testClear(t *testing.T) {
	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Get("http://127.0.0.1:8080/api/v1/addressbook/clear")
	if err != nil {
		t.Logf("Unexpected error: %v", err)
		t.FailNow()
	}
	if resp.StatusCode != http.StatusOK {
		t.Logf("Expected code 200 but got %d", resp.StatusCode)
		t.Fail()
	}
	clearUsersList()
}

func testCreateUser(t *testing.T) {
	client := &http.Client{Timeout: time.Second * 5}
	for i := 0; i < 5; i++ {
		user := &models.User{
			FirstName: fmt.Sprintf("First%d", i),
			LastName:  fmt.Sprintf("Last%d", i),
			Email:     fmt.Sprintf("Email%d", i),
			Phone:     fmt.Sprintf("Phone%d", i),
		}

		data, err := json.Marshal(user)
		if err != nil {
			t.Fatalf("Can't marshal data. Reason: %v", err)
		}
		req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/addressbook/user", bytes.NewReader(data))

		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Can't proceed a request. Reason: %v", err)
			t.FailNow()
		}

		if resp.StatusCode != http.StatusOK {
			t.Logf("Expected code 200, got %d", resp.StatusCode)
			t.FailNow()
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			t.Logf("Unexpected error %v", err)
			t.FailNow()
		}
		resp.Body.Close()
		userslist = append(userslist, *user)
	}
}

func testListUsers(t *testing.T) {
	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Get("http://127.0.0.1:8080/api/v1/addressbook/")
	if err != nil {
		t.Logf("Unexpected error: %v", err)
		t.FailNow()
	}
	if resp.StatusCode != http.StatusOK {
		t.Logf("Expected code 200, got: %d", resp.StatusCode)
	}
	var userlist []models.User
	if err := json.NewDecoder(resp.Body).Decode(&userlist); err != nil {
		t.Logf("Unexpected error: %v", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(userlist, userslist) {
		t.Logf("Arrays doesn't equal.\nHave: %v\n\nWant:%v", userlist, userslist)
		t.FailNow()
	}
}

func testSelectUser(t *testing.T) {
	client := &http.Client{Timeout: time.Second * 5}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:8080/api/v1/addressbook/user/%s", userslist[0].ID.Hex()))
	if err != nil {
		t.FailNow()
	}
	if resp.StatusCode != http.StatusOK {
		t.FailNow()
	}
	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(user, userslist[0]) {
		t.FailNow()
	}
}

func testUpdateUser(t *testing.T) {
	client := &http.Client{Timeout: time.Second * 5}
	userslist[0].FirstName = "updFirst"
	userslist[0].LastName = "updLast"
	userslist[0].Email = "updEmail"
	userslist[0].Phone = "updPhone"
	user := userslist[0]
	data, _ := json.Marshal(&user)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("http://127.0.0.1:8080/api/v1/addressbook/user/%s", user.ID.Hex()), bytes.NewReader(data))
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Can't proceed a request. Reason: %v", err)
		t.FailNow()
	}
	if resp.StatusCode != http.StatusNoContent {
		t.FailNow()
	}
}

func testRemoveUser(t *testing.T) {
	client := &http.Client{Timeout: time.Second * 5}
	user := userslist[1]
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://127.0.0.1:8080/api/v1/addressbook/user/%s", user.ID.Hex()), nil)
	if err != nil {
		t.FailNow()
	}
	resp, err := client.Do(req)
	if err != nil {
		t.FailNow()
	}
	if resp.StatusCode != http.StatusNoContent {
		t.FailNow()
	}
	removeUser(user.ID)
}
