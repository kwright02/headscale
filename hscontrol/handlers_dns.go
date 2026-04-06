package hscontrol

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/juanfont/headscale/hscontrol/types"
	"github.com/rs/zerolog/log"
	"tailscale.com/util/dnsname"
)

// dnsRecordResponse is the JSON representation of a DNSRecord returned by the API.
type dnsRecordResponse struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

func dnsRecordToResponse(r types.DNSRecord) dnsRecordResponse {
	return dnsRecordResponse{
		ID:        r.ID,
		Name:      r.Name,
		Type:      r.Type,
		Value:     r.Value,
		CreatedAt: r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// ListDNSRecords handles GET /api/v1/dns/records.
// It returns all custom A/AAAA DNS records.
func (h *Headscale) ListDNSRecords(w http.ResponseWriter, r *http.Request) {
	records, err := h.state.ListDNSRecords()
	if err != nil {
		log.Error().Caller().Err(err).Msg("listing DNS records")
		httpError(w, NewHTTPError(http.StatusInternalServerError, "failed to list DNS records", err))

		return
	}

	resp := make([]dnsRecordResponse, len(records))
	for i, rec := range records {
		resp[i] = dnsRecordToResponse(rec)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]any{"records": resp}); err != nil {
		log.Error().Caller().Err(err).Msg("encoding DNS records response")
	}
}

// createDNSRecordRequest is the request body for creating a DNS record.
type createDNSRecordRequest struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CreateDNSRecord handles POST /api/v1/dns/records.
// It validates and creates a new A or AAAA DNS record.
func (h *Headscale) CreateDNSRecord(w http.ResponseWriter, r *http.Request) {
	var req createDNSRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, NewHTTPError(http.StatusBadRequest, "invalid request body", err))

		return
	}

	if req.Name == "" {
		httpError(w, NewHTTPError(http.StatusBadRequest, "name is required", nil))

		return
	}

	if req.Type == "" {
		httpError(w, NewHTTPError(http.StatusBadRequest, "type is required", nil))

		return
	}

	if req.Value == "" {
		httpError(w, NewHTTPError(http.StatusBadRequest, "value is required", nil))

		return
	}

	// Validate name as a proper DNS name (single label or FQDN).
	normalizedName, err := validateDNSRecordName(req.Name)
	if err != nil {
		httpError(w, NewHTTPError(http.StatusBadRequest, err.Error(), err))

		return
	}

	// Validate type and value.
	if err := types.ValidateDNSRecord(req.Type, req.Value); err != nil {
		code := http.StatusBadRequest
		if errors.Is(err, types.ErrDNSRecordInvalidType) {
			code = http.StatusUnprocessableEntity
		}
		httpError(w, NewHTTPError(code, err.Error(), err))

		return
	}

	record, c, err := h.state.CreateDNSRecord(h.cfg, normalizedName, req.Type, req.Value)
	if err != nil {
		if errors.Is(err, types.ErrDNSRecordAlreadyExists) {
			httpError(w, NewHTTPError(http.StatusConflict, "DNS record already exists", err))

			return
		}

		log.Error().Caller().Err(err).Msg("creating DNS record")
		httpError(w, NewHTTPError(http.StatusInternalServerError, "failed to create DNS record", err))

		return
	}

	h.Change(c)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]any{"record": dnsRecordToResponse(*record)}); err != nil {
		log.Error().Caller().Err(err).Msg("encoding create DNS record response")
	}
}

// DeleteDNSRecord handles DELETE /api/v1/dns/records/{id}.
// It removes a DNS record by its numeric ID.
func (h *Headscale) DeleteDNSRecord(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httpError(w, NewHTTPError(http.StatusBadRequest, "invalid record id", err))

		return
	}

	c, err := h.state.DeleteDNSRecord(h.cfg, id)
	if err != nil {
		if errors.Is(err, types.ErrDNSRecordNotFound) {
			httpError(w, NewHTTPError(http.StatusNotFound, "DNS record not found", err))

			return
		}

		log.Error().Caller().Err(err).Msg("deleting DNS record")
		httpError(w, NewHTTPError(http.StatusInternalServerError, "failed to delete DNS record", err))

		return
	}

	h.Change(c)

	w.WriteHeader(http.StatusNoContent)
}

// validateDNSRecordName validates a DNS name for storage.
// It accepts both single-label hostnames (e.g. "web") and fully-qualified domain
// names (e.g. "grafana.myvpn.example.com" or "grafana.myvpn.example.com.").
// The name is stored as provided by the caller; no normalization is applied.
func validateDNSRecordName(name string) (string, error) {
	if name == "" {
		return "", types.ErrDNSRecordInvalidName
	}

	_, err := dnsname.ToFQDN(name)
	if err != nil {
		return "", types.ErrDNSRecordInvalidName
	}

	return name, nil
}
