// Package schema embeds the SQL migrations so the binary can apply them on startup.
package schema

import "embed"

// Migrations holds the golang-migrate style files: NNNNNN_name.up.sql and NNNNNN_name.down.sql.
//
//go:embed migrations/*.sql
var Migrations embed.FS
