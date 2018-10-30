package db

import (
	"errors"
	"io"
	"log"
	"time"

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
	DB *mgo.Database
}

var logger *log.Logger

// NewV2 updated version of constructor.
func NewV2(dbconf types.DB) (*Repo, error) {
	repo, err := mgo.Dial(dbconf.Connection)
	if err != nil {
		return &Repo{}, err
	}
	var r Repo
	r.DB = repo.DB(dbconf.Name)
	return &r, nil
}

//New makes a new config file for start database connection
// will be removed soon
func New(_, _, _, _, _ string, _ time.Duration, _ int) (*Config, error) {

	return &Config{}, nil
}

//CreateConnection creates connection to database and returns connection and error.
func CreateConnection(c *Config, w io.Writer) (*mgo.Collection, error) {
	logger = log.New(w, "[Database] ", log.Lshortfile+log.Ldate+log.Ltime)
	logger.Printf("Opening connection to database: %s", c.Database)
	session, err := func(tries int) (*mgo.Session, error) {
		for tries > 0 {
			session, err := mgo.DialWithTimeout(c.ConnString, c.Timeout)
			if err != nil {
				tries--
				logger.Printf("Can't connect to mongoDB. Reason: %v", err)
				logger.Printf("Reconnecting in 5 seconds (Attempt(s) left: %d)\n\n", tries)
				time.Sleep(time.Second * 3)
				continue
			}
			return session, nil
		}
		return nil, errors.New("Got a problem connecting to db. Check logs")
	}(c.TryAmount)
	if err != nil {
		return nil, err
	}
	logger.Println("Connection to mongoDB was successful")
	go checkConnection(session, w)
	return session.DB(c.Database).C(c.Collection), nil
}

func checkConnection(session *mgo.Session, w io.Writer) {
	logger := log.New(w, "[Database Connection Checker] ", log.Lshortfile+log.Ldate+log.Ltime)
	ticker := time.NewTicker(time.Second * 5)
	lostConnection := false
	for range ticker.C {
		if err := session.Ping(); err != nil {
			logger.Println("Lost connection to database. Trying to reconnect")
			lostConnection = true
			session.Refresh()
			continue
		}
		if lostConnection {
			lostConnection = false
			logger.Println("Connection to database has been restored")
		}
	}
}
