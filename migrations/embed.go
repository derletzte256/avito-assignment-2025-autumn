package migrations

import (
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

func init() {
	goose.SetBaseFS(embedMigrations)
}
