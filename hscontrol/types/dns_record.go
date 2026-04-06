package types

import (
	"errors"
	"fmt"
	"net/netip"
	"time"
)

// DNSRecordType represents a supported DNS record type.
type DNSRecordType string

const (
	DNSRecordTypeA    DNSRecordType = "A"
	DNSRecordTypeAAAA DNSRecordType = "AAAA"
)

var (
	// ErrDNSRecordInvalidType is returned when an unsupported DNS record type is specified.
	ErrDNSRecordInvalidType = errors.New("invalid DNS record type: only A and AAAA are supported")
	// ErrDNSRecordInvalidValue is returned when the IP address is invalid for the record type.
	ErrDNSRecordInvalidValue = errors.New("invalid IP address for DNS record type")
	// ErrDNSRecordNotFound is returned when a DNS record cannot be found.
	ErrDNSRecordNotFound = errors.New("DNS record not found")
	// ErrDNSRecordAlreadyExists is returned when a duplicate DNS record is detected.
	ErrDNSRecordAlreadyExists = errors.New("DNS record already exists")
	// ErrDNSRecordInvalidName is returned when the record name is not a valid DNS name.
	ErrDNSRecordInvalidName = errors.New("invalid DNS record name")
)

// DNSRecord represents a custom A/AAAA DNS record stored in the database.
type DNSRecord struct {
	ID    uint64 `gorm:"primaryKey;autoIncrement"`
	Name  string `gorm:"not null;uniqueIndex:idx_dns_records_unique"`
	Type  string `gorm:"not null;uniqueIndex:idx_dns_records_unique"` // "A" or "AAAA"
	Value string `gorm:"not null;uniqueIndex:idx_dns_records_unique"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ValidateDNSRecord validates the type and value of a DNS record before persistence.
// Returns an error if the record type is not A/AAAA, or if the IP address is
// inconsistent with the record type.
func ValidateDNSRecord(recordType, value string) error {
	switch DNSRecordType(recordType) {
	case DNSRecordTypeA:
		addr, err := netip.ParseAddr(value)
		if err != nil || !addr.Is4() {
			return fmt.Errorf("%w: A record requires a valid IPv4 address, got %q", ErrDNSRecordInvalidValue, value)
		}
	case DNSRecordTypeAAAA:
		addr, err := netip.ParseAddr(value)
		if err != nil || !addr.Is6() {
			return fmt.Errorf("%w: AAAA record requires a valid IPv6 address, got %q", ErrDNSRecordInvalidValue, value)
		}
	default:
		return fmt.Errorf("%w: got %q", ErrDNSRecordInvalidType, recordType)
	}

	return nil
}
