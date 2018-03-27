package controllers

import (
	"reflect"
	"testing"

	"github.com/ferux/AddressBook/models"
	"gopkg.in/mgo.v2/bson"
)

func TestUser_CreateUser(t *testing.T) {
	type args struct {
		u *models.User
	}
	tests := []struct {
		name    string
		c       *User
		args    args
		want    bson.ObjectId
		wantErr bool
	}{
		// TODO: Add test cases. id:4 gh:5 ic:gh
		{
			name: "Nil pointer test",
			c: &User{
				Collection: nil,
			},
			args: args{
				u: nil,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.CreateUser(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_UpdateUser(t *testing.T) {
	type args struct {
		u *models.User
	}
	tests := []struct {
		name    string
		c       *User
		args    args
		wantErr bool
	}{
	// TODO: Add test cases. id:0 gh:1 ic:gh
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.UpdateUser(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("User.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_DeleteUser(t *testing.T) {
	type args struct {
		id bson.ObjectId
	}
	tests := []struct {
		name    string
		c       *User
		args    args
		wantErr bool
	}{
	// TODO: Add test cases. id:2 gh:3 ic:gh
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.DeleteUser(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("User.DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_SelectUser(t *testing.T) {
	type args struct {
		id bson.ObjectId
	}
	tests := []struct {
		name    string
		c       *User
		args    args
		want    *models.User
		wantErr bool
	}{
	// TODO: Add test cases. id:3 gh:4 ic:gh
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.SelectUser(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.SelectUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.SelectUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_ListUsers(t *testing.T) {
	tests := []struct {
		name    string
		c       *User
		want    *[]models.User
		wantErr bool
	}{
	// TODO: Add test cases. id:6 gh:7 ic:gh
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.ListUsers()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.ListUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}
