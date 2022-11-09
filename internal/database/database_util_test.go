package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO : REMOVE (temporary for migration)
func TestConsole(t *testing.T) {
	t.Skip()
	dsn := "host=localhost user=root password=password dbname=kek port=9920 sslmode=disable TimeZone=Asia/Tokyo"
	err := migrateDB(dsn, "/migrations")
	assert.NoError(t, err)
}
