package main

import (
	"github.com/ferux/addressbook/internal/api"
	"github.com/ferux/addressbook/internal/db"
	"github.com/ferux/addressbook/internal/types"
)

func run(c *types.Config) error {
	repo, err := db.New(c.Database)
	if err != nil {
		return err
	}
	api := api.NewAPI(repo, c.API)
	return api.Run()
}
