package types

import "time"

// Config is an app-wode configuration
type Config struct {
	Database     DB   `json:"database"`
	DatabaseTest DB   `json:"database_test"`
	API          API  `json:"api"`
	Debug        bool `json:"debug"`
	CustomTestDB bool `json:"custom_test_db"`
}

// DB is a configuration of DB
type DB struct {
	Connection string `json:"connection,omitempty"`
	Name       string `json:"name,omitempty"`
}

// API is a configuration of API
type API struct {
	Listen string        `json:"listen,omitempty"`
	Timeot time.Duration `json:"timeout,omitempty"`
}
