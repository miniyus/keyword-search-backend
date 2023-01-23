package main

import (
	"github.com/miniyus/keyword-search-backend/config"
	"github.com/miniyus/keyword-search-backend/database"
	"github.com/miniyus/keyword-search-backend/database/migrations"
)

func main() {
	configure := config.GetConfigs()

	db := database.DB(configure.Database)

	migrations.Migrate(db)
}
