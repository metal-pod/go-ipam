package ipam

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func createDB(t *testing.T) (*sql, error) {
	dbname := "postgres"
	db, err := NewPostgresStorage("localhost", "5433", "postgres", "password", dbname, "disable")
	err = db.db.Ping()
	require.Nil(t, err)
	return db, err
}

func destroy(s *sql) {
	s.db.MustExec("DROP TABLE prefixes")
}

func Test_sql_prefixExists(t *testing.T) {
	db, err := createDB(t)
	require.Nil(t, err)
	require.NotNil(t, db)

	// Existing Prefix
	prefix := &Prefix{Cidr: "10.0.0.0/16"}
	p, err := db.CreatePrefix(prefix)
	require.Nil(t, err)
	require.NotNil(t, p)
	require.Equal(t, prefix.Cidr, p.Cidr)
	got, exists := db.prefixExists(prefix)
	require.True(t, exists)
	require.Equal(t, got.Cidr, prefix.Cidr)

	// NonExisting Prefix
	notExistingPrefix := &Prefix{Cidr: "10.0.0.0/8"}
	got, exists = db.prefixExists(notExistingPrefix)
	require.False(t, exists)
	require.Nil(t, got)

	// Delete Existing Prefix
	_, err = db.DeletePrefix(prefix)
	got, exists = db.prefixExists(prefix)
	require.False(t, exists)
	require.Nil(t, got)

	// cleanup
	destroy(db)
}

func Test_sql_CreatePrefix(t *testing.T) {
	db, err := createDB(t)
	require.Nil(t, err)
	require.NotNil(t, db)

	// Existing Prefix
	prefix := &Prefix{Cidr: "11.0.0.0/16"}
	got, exists := db.prefixExists(prefix)
	require.False(t, exists)
	require.Nil(t, got)
	p, err := db.CreatePrefix(prefix)
	require.Nil(t, err)
	require.NotNil(t, p)
	require.Equal(t, prefix.Cidr, p.Cidr)
	got, exists = db.prefixExists(prefix)
	require.True(t, exists)
	require.Equal(t, got.Cidr, prefix.Cidr)

	// Duplicate Prefix
	p, err = db.CreatePrefix(prefix)
	require.Nil(t, err)
	require.NotNil(t, p)
	require.Equal(t, prefix.Cidr, p.Cidr)

	ps, err := db.ReadAllPrefixes()
	require.Nil(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 1, len(ps))

	// cleanup
	destroy(db)
}

func Test_sql_ReadPrefix(t *testing.T) {
	db, err := createDB(t)
	require.Nil(t, err)
	require.NotNil(t, db)

	// Prefix
	p, err := db.ReadPrefix("12.0.0.0/8")
	require.NotNil(t, err)
	require.Equal(t, "unable to read prefix:sql: no rows in result set", err.Error())
	require.Nil(t, p)

	prefix := &Prefix{Cidr: "12.0.0.0/16"}
	p, err = db.CreatePrefix(prefix)
	require.Nil(t, err)
	require.NotNil(t, p)

	p, err = db.ReadPrefix("12.0.0.0/16")
	require.Nil(t, err)
	require.NotNil(t, p)
	require.Equal(t, "12.0.0.0/16", p.Cidr)

	// cleanup
	destroy(db)
}

func Test_sql_ReadAllPrefix(t *testing.T) {
	db, err := createDB(t)
	require.Nil(t, err)
	require.NotNil(t, db)

	// no Prefixes
	ps, err := db.ReadAllPrefixes()
	require.Nil(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 0, len(ps))

	// One Prefix
	prefix := &Prefix{Cidr: "12.0.0.0/16"}
	p, err := db.CreatePrefix(prefix)
	require.Nil(t, err)
	require.NotNil(t, p)
	ps, err = db.ReadAllPrefixes()
	require.Nil(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 1, len(ps))

	// no Prefixes again
	_, err = db.DeletePrefix(prefix)
	require.Nil(t, err)
	ps, err = db.ReadAllPrefixes()
	require.Nil(t, err)
	require.NotNil(t, ps)
	require.Equal(t, 0, len(ps))

	// cleanup
	destroy(db)
}

func Test_sql_UpdatePrefix(t *testing.T) {
	db, err := createDB(t)
	require.Nil(t, err)
	require.NotNil(t, db)

	// Prefix
	prefix := &Prefix{Cidr: "13.0.0.0/16", ParentCidr: "13.0.0.0/8"}
	p, err := db.CreatePrefix(prefix)
	require.Nil(t, err)
	require.NotNil(t, p)

	// Check if present
	p, err = db.ReadPrefix("13.0.0.0/16")
	require.Nil(t, err)
	require.NotNil(t, p)
	require.Equal(t, "13.0.0.0/16", p.Cidr)
	require.Equal(t, "13.0.0.0/8", p.ParentCidr)

	// Modify
	prefix.ParentCidr = "13.0.0.0/12"
	p, err = db.UpdatePrefix(prefix)
	p, err = db.ReadPrefix("13.0.0.0/16")
	require.Nil(t, err)
	require.NotNil(t, p)
	require.Equal(t, "13.0.0.0/16", p.Cidr)
	require.Equal(t, "13.0.0.0/12", p.ParentCidr)

	// cleanup
	destroy(db)
}