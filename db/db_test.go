package db

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
)

const testConnectionString = "mongodb://192.168.99.100:27017"

func TestNew(t *testing.T) {
	type args struct {
		host       string
		user       string
		password   string
		database   string
		collection string
		timeout    time.Duration
		tryAmount  int
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test empty host",
			args: args{
				host:       "",
				user:       "1",
				password:   "2",
				database:   "3",
				collection: "4",
				timeout:    time.Second,
				tryAmount:  1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test empty database",
			args: args{
				host:       "1",
				user:       "2",
				password:   "3",
				database:   "",
				collection: "4",
				timeout:    time.Second,
				tryAmount:  1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test empty collection",
			args: args{
				host:       "1",
				user:       "2",
				password:   "3",
				database:   "4",
				collection: "",
				timeout:    time.Second,
				tryAmount:  1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test negative timeout",
			args: args{
				host:       "1",
				user:       "1",
				password:   "2",
				database:   "3",
				collection: "4",
				timeout:    -1 * time.Second,
				tryAmount:  1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test negative tryAmount",
			args: args{
				host:       "1",
				user:       "1",
				password:   "2",
				database:   "3",
				collection: "4",
				timeout:    1 * time.Second,
				tryAmount:  -1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Test zero timeOut and tryAmount",
			args: args{
				host:       "1",
				user:       "1",
				password:   "2",
				database:   "3",
				collection: "4",
				timeout:    0,
				tryAmount:  0,
			},
			want: &Config{
				ConnString: "mongodb://1:2@1/3",
				Collection: "4",
				Database:   "3",
				Timeout:    time.Second * 5,
				TryAmount:  1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.host, tt.args.user, tt.args.password, tt.args.database, tt.args.collection, tt.args.timeout, tt.args.tryAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateConnection(t *testing.T) {
	conf, err := New("192.168.99.101:27017", "", "", "test", "test", time.Second, 1)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	coll, err := CreateConnection(conf, ioutil.Discard)
	if err == nil || coll != nil {
		t.Errorf("Expected to got an error and coll should be nil. Got: %v %v", err, coll)
		t.Fail()
	}
	conf.ConnString = testConnectionString
	coll, err = CreateConnection(conf, ioutil.Discard)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
		t.Fail()
	}
	if coll == nil {
		t.Error("Coll shouldn't be empty")
		t.Fail()
	}
	defer coll.Database.Session.Close()
}

func Test_checkConnection(t *testing.T) {
	type args struct {
		session *mgo.Session
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			checkConnection(tt.args.session, w)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("checkConnection() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
