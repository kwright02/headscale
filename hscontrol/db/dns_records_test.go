package db

import (
	"testing"

	"github.com/juanfont/headscale/hscontrol/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDNSRecord(t *testing.T) {
	db, err := newSQLiteTestDB()
	require.NoError(t, err)

	rec, err := db.CreateDNSRecord("web.example.com.", "A", "192.0.2.1")
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.NotZero(t, rec.ID)
	assert.Equal(t, "web.example.com.", rec.Name)
	assert.Equal(t, "A", rec.Type)
	assert.Equal(t, "192.0.2.1", rec.Value)
}

func TestCreateDNSRecordAAAA(t *testing.T) {
	db, err := newSQLiteTestDB()
	require.NoError(t, err)

	rec, err := db.CreateDNSRecord("ipv6.example.com.", "AAAA", "2001:db8::1")
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Equal(t, "AAAA", rec.Type)
	assert.Equal(t, "2001:db8::1", rec.Value)
}

func TestCreateDNSRecordDuplicate(t *testing.T) {
	db, err := newSQLiteTestDB()
	require.NoError(t, err)

	_, err = db.CreateDNSRecord("dup.example.com.", "A", "10.0.0.1")
	require.NoError(t, err)

	// Same (name, type, value) -> conflict.
	_, err = db.CreateDNSRecord("dup.example.com.", "A", "10.0.0.1")
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrDNSRecordAlreadyExists)
}

func TestCreateDNSRecordSameNameDifferentType(t *testing.T) {
	db, err := newSQLiteTestDB()
	require.NoError(t, err)

	_, err = db.CreateDNSRecord("dual.example.com.", "A", "192.0.2.2")
	require.NoError(t, err)

	// Same name but different type is allowed.
	_, err = db.CreateDNSRecord("dual.example.com.", "AAAA", "2001:db8::2")
	require.NoError(t, err)
}

func TestListDNSRecords(t *testing.T) {
	db, err := newSQLiteTestDB()
	require.NoError(t, err)

	// Empty list.
	records, err := db.ListDNSRecords()
	require.NoError(t, err)
	assert.Empty(t, records)

	_, err = db.CreateDNSRecord("a.example.com.", "A", "1.2.3.4")
	require.NoError(t, err)

	_, err = db.CreateDNSRecord("b.example.com.", "AAAA", "::1")
	require.NoError(t, err)

	records, err = db.ListDNSRecords()
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestDeleteDNSRecord(t *testing.T) {
	db, err := newSQLiteTestDB()
	require.NoError(t, err)

	rec, err := db.CreateDNSRecord("del.example.com.", "A", "5.6.7.8")
	require.NoError(t, err)

	err = db.DeleteDNSRecord(rec.ID)
	require.NoError(t, err)

	records, err := db.ListDNSRecords()
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestDeleteDNSRecordNotFound(t *testing.T) {
	db, err := newSQLiteTestDB()
	require.NoError(t, err)

	err = db.DeleteDNSRecord(9999)
	require.Error(t, err)
	assert.ErrorIs(t, err, types.ErrDNSRecordNotFound)
}
