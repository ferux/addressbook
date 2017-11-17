package db

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

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

var logger *log.Logger

//New makes a new config file for start database connection
func New(host, user, password, database, collection string, timeout time.Duration, tryAmount int) (*Config, error) {
	if len(database) == 0 || len(host) == 0 || len(collection) == 0 {
		return nil, errors.New("Parameters shouldn't be empty")
	}
	if timeout < 0 || tryAmount < 0 {
		return nil, errors.New("timeout or amount of tries should be greater or equal zero")
	}
	if timeout == 0 {
		timeout = time.Second * 5
	}
	if tryAmount == 0 {
		tryAmount = 1
	}
	creds := ""
	if len(user) > 0 {
		creds = fmt.Sprintf("%s:%s@", user, password)
	}
	config := Config{}
	config.ConnString = fmt.Sprintf("mongodb://%s%s/%s", creds, host, database)
	config.Database = database
	config.Collection = collection
	config.Timeout = timeout
	config.TryAmount = tryAmount
	return &config, nil
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
