package db

import (
	"time"

	"github.com/ferux/addressbook"

	"github.com/sirupsen/logrus"

	"github.com/ferux/addressbook/internal/types"

	"gopkg.in/mgo.v2"
)

//Config is a struct to database configure
type Config struct {
	Database   string
	ConnString string
	Collection string
	TryAmount  int
	Timeout    time.Duration
}

// Repo contains active connection to mgo.
type Repo struct {
	Session *mgo.Session
	DB      *mgo.Database
	conf    types.DB
	logger  *logrus.Entry
	status  addressbook.Code
}

// New updated version of constructor.
func New(dbconf types.DB) (*Repo, error) {
	r := Repo{
		conf: dbconf,
		logger: logrus.New().WithFields(logrus.Fields{
			"package": "db",
			"entity":  "repo",
		}),
		status: addressbook.Unknown,
	}

	err := r.connect()
	go r.keepConnection()
	return &r, err
}

func (r *Repo) keepConnection() {
	logger := r.logger.WithField("fn", "keepConnection")
	for {
		err := r.Session.Ping()
		if err != nil {
			r.status = addressbook.HaveProblems
			logger.WithError(err).Error("can't ping database")
			r.Session.Refresh()
			time.Sleep(time.Second * 3)
			continue
		}
		r.status = addressbook.Running
		time.Sleep(time.Second * 10)
	}
}

func (r *Repo) connect() (err error) {
	if r.Session != nil {
		r.Session.Close()
		r.Session = nil
		r.DB = nil
	}

	r.Session, err = mgo.Dial(r.conf.Connection)
	if err != nil {
		return err
	}
	r.Session.SetSocketTimeout(time.Second * 15)
	r.Session.SetSyncTimeout(time.Second * 15)
	r.DB = r.Session.DB(r.conf.Name)
	return nil
}

// GetStatus returns current status
func (r *Repo) GetStatus() addressbook.Code {
	return r.status
}
