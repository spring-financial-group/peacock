package server

import "github.com/spring-financial-group/peacock/pkg/config"

type DataSources struct {
}

// initDataSources creates the data sources for the server, initialising Redis, Postgres, etc.
func initDataSources(cfg *config.Config) (*DataSources, error) {
	return &DataSources{}, nil
}
