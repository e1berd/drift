package migrate

import (
	"strings"
	"testing"
	"testing/fstest"

	schema "github.com/e1berd/drift/app/db"
)

func TestCollectRejectsDuplicateVersions(t *testing.T) {
	fsys := fstest.MapFS{
		"m/000007_a.up.sql": {Data: []byte("select 1")},
		"m/000007_b.up.sql": {Data: []byte("select 1")},
	}
	_, err := collect(fsys, "m")
	if err == nil || !strings.Contains(err.Error(), "version 7") {
		t.Fatalf("want duplicate version error, got %v", err)
	}
}

func TestEmbeddedMigrationsHaveUniqueVersions(t *testing.T) {
	list, err := collect(schema.Migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.Fatal("no embedded migrations")
	}
}
