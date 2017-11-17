package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ferux/AddressBook/controllers"
	"github.com/ferux/AddressBook/daemon"
	"github.com/ferux/AddressBook/db"
)

//VERSION Shows the version of the application
const VERSION = "1.0"

var logger *log.Logger
var controller controllers.User

//Parameters for application
var username, password, host, database, collection, listen string
var timeout, retry int

//debugmode sets server to debug mode
var debugmode = false

func getParams() {
	flag.StringVar(&username, "dbuser", "", "Database username")
	flag.StringVar(&password, "dbpass", "", "Database username's password")
	flag.StringVar(&host, "dbaddr", "127.0.0.1:27017", "Database Address in format Address:Port")
	flag.StringVar(&database, "dbname", "", "Database name")
	flag.StringVar(&collection, "dbcoll", "", "Collection where server will store records")
	flag.IntVar(&timeout, "timeout", 3, "Connection to database timeout")
	flag.IntVar(&retry, "retry", 3, "Number of tries before exit if the connection to database fails")
	flag.StringVar(&listen, "listen", ":8080", "API listen address")
	flag.BoolVar(&debugmode, "debug", false, "Puts application into debug mode")
	flag.Parse()
}

func main() {
	getParams()

	w := ioutil.Discard
	if debugmode {
		w = os.Stderr
	}

	logger = log.New(w, "[Addressbook] ", log.Ldate+log.Ltime)
	logger.Printf("Started Addressbook. Version: %s", VERSION)

	// dbConfig, err := db.New(host, username, password, database, collection, time.Second*time.Duration(timeout), retry)
	dbConfig, err := db.New(host, username, password, database, collection, time.Second*time.Duration(timeout), retry)
	if err != nil {
		log.Fatalf("Failed to parse parameters. Reason: %v", err)
		os.Exit(2)
	}
	collection, err := db.CreateConnection(dbConfig, w)
	if err != nil {
		log.Fatalf("Can't connect to database. Reason: %v", err)
		os.Exit(2)
	}
	controller = controllers.User{Collection: collection}
	if err := daemon.Start(collection, ":8080", w); err != nil {
		os.Exit(2)
	}
}
