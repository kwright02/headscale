package db

import (
	"fmt"
	"strings"

	"github.com/juanfont/headscale/hscontrol/types"
)

// ListDNSRecords returns all DNS records from the database.
func (hsdb *HSDatabase) ListDNSRecords() ([]types.DNSRecord, error) {
	var records []types.DNSRecord

	if err := hsdb.DB.Order("id").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("listing DNS records: %w", err)
	}

	return records, nil
}

// CreateDNSRecord creates a new DNS record in the database.
// Returns ErrDNSRecordAlreadyExists if a record with the same name, type, and value exists.
func (hsdb *HSDatabase) CreateDNSRecord(name, recordType, value string) (*types.DNSRecord, error) {
	record := types.DNSRecord{
		Name:  name,
		Type:  recordType,
		Value: value,
	}

	if err := hsdb.DB.Create(&record).Error; err != nil {
		if isUniqueConstraintViolation(err) {
			return nil, types.ErrDNSRecordAlreadyExists
		}

		return nil, fmt.Errorf("creating DNS record: %w", err)
	}

	return &record, nil
}

// DeleteDNSRecord removes a DNS record by its ID.
// Returns ErrDNSRecordNotFound if no record with the given ID exists.
func (hsdb *HSDatabase) DeleteDNSRecord(id uint64) error {
	result := hsdb.DB.Delete(&types.DNSRecord{}, id)
	if result.Error != nil {
		return fmt.Errorf("deleting DNS record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return types.ErrDNSRecordNotFound
	}

	return nil
}

// isUniqueConstraintViolation returns true when err is a unique-index violation
// from SQLite ("UNIQUE constraint failed") or PostgreSQL ("duplicate key value").
func isUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "duplicate key value")
}
