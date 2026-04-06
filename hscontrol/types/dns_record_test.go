package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDNSRecord(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
		value      string
		wantErr    bool
		errIs      error
	}{
		{
			name:       "valid A record",
			recordType: "A",
			value:      "192.0.2.1",
			wantErr:    false,
		},
		{
			name:       "valid AAAA record",
			recordType: "AAAA",
			value:      "2001:db8::1",
			wantErr:    false,
		},
		{
			name:       "A record with IPv6 address",
			recordType: "A",
			value:      "2001:db8::1",
			wantErr:    true,
			errIs:      ErrDNSRecordInvalidValue,
		},
		{
			name:       "AAAA record with IPv4 address",
			recordType: "AAAA",
			value:      "192.0.2.1",
			wantErr:    true,
			errIs:      ErrDNSRecordInvalidValue,
		},
		{
			name:       "unsupported record type MX",
			recordType: "MX",
			value:      "mail.example.com",
			wantErr:    true,
			errIs:      ErrDNSRecordInvalidType,
		},
		{
			name:       "unsupported record type CNAME",
			recordType: "CNAME",
			value:      "alias.example.com",
			wantErr:    true,
			errIs:      ErrDNSRecordInvalidType,
		},
		{
			name:       "empty type",
			recordType: "",
			value:      "1.2.3.4",
			wantErr:    true,
			errIs:      ErrDNSRecordInvalidType,
		},
		{
			name:       "invalid IPv4 address for A record",
			recordType: "A",
			value:      "not-an-ip",
			wantErr:    true,
			errIs:      ErrDNSRecordInvalidValue,
		},
		{
			name:       "invalid IPv6 address for AAAA record",
			recordType: "AAAA",
			value:      "not-an-ip",
			wantErr:    true,
			errIs:      ErrDNSRecordInvalidValue,
		},
		{
			name:       "A record with valid IPv4-mapped IPv6",
			recordType: "A",
			value:      "::ffff:192.0.2.1",
			wantErr:    true, // Is6() is true for IPv4-mapped IPv6
			errIs:      ErrDNSRecordInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDNSRecord(tt.recordType, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errIs != nil {
					assert.ErrorIs(t, err, tt.errIs)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
