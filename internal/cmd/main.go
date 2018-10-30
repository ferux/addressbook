package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ferux/addressbook"
	"github.com/ferux/addressbook/internal/types"
)

var logger *log.Logger

func loadConfig() *types.Config {
	conf := types.Config{}
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalf("can't read config file: %v", err)
	}
	if err = json.Unmarshal(data, &conf); err != nil {
		log.Fatalf("can't unmarshal data: %v", err)
	}
	return &conf
}

func main() {
	conf := loadConfig()
	w := ioutil.Discard
	if conf.Debug {
		w = os.Stdout
	}
	logger = log.New(w, "AddressBook ", log.Ltime)
	logger.Printf("started app ver=%s rev=%s env=%s", addressbook.Version, addressbook.Revision, addressbook.Env)
	logger.Printf("using conf=%+v", conf)
	if err := run(conf); err != nil {
		log.Println("err running app: ", err)
	}
	defer logger.Println("finished")
}
