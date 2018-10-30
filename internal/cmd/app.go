package main

import (
	"os"

	"github.com/ferux/addressbook/internal/daemon"
	"github.com/ferux/addressbook/internal/db"
	"github.com/ferux/addressbook/internal/types"
)

func run(c *types.Config) error {
	repo, err := db.NewV2(c.Database)
	if err != nil {
		return err
	}

	return daemon.Start(repo.DB, c.API.Listen, os.Stdout)
}
