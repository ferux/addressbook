package main

import (
	"github.com/ferux/addressbook/internal/daemon"
	"github.com/ferux/addressbook/internal/db"
	"github.com/ferux/addressbook/internal/types"
)

func run(c *types.Config) error {
	repo, err := db.NewV2(c.Database)
	if err != nil {
		return err
	}
	api := daemon.NewAPI(repo.DB, c.API)
	return api.Run()
}
