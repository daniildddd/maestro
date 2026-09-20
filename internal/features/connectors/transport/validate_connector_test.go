package transport_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/logger"
	core_http_response "github.com/daniildddd/maestro/internal/core/transport/response"
	"github.com/daniildddd/maestro/internal/features/connectors/transport"
)

func newValidateRequest(t *testing.T, body, contentType string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req.WithContext(logger.ToContext(req.Context(), nopLogger()))
}

func TestConnectorsHTTPHandler_ValidateConnector(t *testing.T) {
	t.Parallel()

	validReport := domain.ValidationReport{
		Valid: true,
		Steps: []domain.ValidationStep{
			{
				ID:     domain.StepConfig,
				Status: domain.CheckSeverityOK,
				Checks: []domain.ValidationCheck{{
					ID:       "config.schema",
					Severity: domain.CheckSeverityOK,
				}},
			},
			{
				ID:     domain.StepTables,
				Status: domain.CheckSeverityWarning,
				Checks: []domain.ValidationCheck{{
					ID:       "postgres.table.primary_key",
					Severity: domain.CheckSeverityWarning,
					Table:    "public.payments",
				}},
			},
		},
	}

	invalidReport := domain.ValidationReport{
		Valid: false,
		Steps: []domain.ValidationStep{{
			ID:     domain.StepCDC,
			Status: domain.CheckSeverityError,
			Checks: []domain.ValidationCheck{{
				ID:       "postgres.cdc.wal_level",
				Severity: domain.CheckSeverityError,
				Message:  "wal_level = 'replica'",
			}},
		}},
	}

	validBody := transport.ValidateResponse{
		Valid: true,
		Steps: []transport.ValidateStepResponse{
			{
				ID:     "config",
				Status: "ok",
				Checks: []transport.ValidateCheckResponse{{ID: "config.schema", Severity: "ok"}},
			},
			{
				ID:     "tables",
				Status: "warning",
				Checks: []transport.ValidateCheckResponse{{
					ID:       "postgres.table.primary_key",
					Severity: "warning",
					Table:    "public.payments",
				}},
			},
		},
	}

	invalidBody := transport.ValidateResponse{
		Valid: false,
		Steps: []transport.ValidateStepResponse{{
			ID:     "cdc",
			Status: "error",
			Checks: []transport.ValidateCheckResponse{{
				ID:       "postgres.cdc.wal_level",
				Severity: "error",
				Message:  "wal_level = 'replica'",
			}},
		}},
	}

	tests := []struct {
		name        string
		body        string
		contentType string
		setupMock   func(m *MockConnectorsService)
		wantStatus  int
		wantCode    string
		wantBody    transport.ValidateResponse
		wantDetails transport.ValidateResponse
	}{
		{
			name: "valid report returns 200 with warnings",
			body: `{
				"plugin_type": "io.debezium.connector.postgresql.PostgresConnector",
				"name": "pg-connector",
				"config": {"database.hostname": "pg"}
			}`,
			contentType: "application/json",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					ValidateConnector(
						mock.Anything,
						"io.debezium.connector.postgresql.PostgresConnector",
						map[string]string{"name": "pg-connector", "database.hostname": "pg"},
					).
					Return(validReport, nil).
					Once()
			},
			wantStatus: http.StatusOK,
			wantBody:   validBody,
		},
		{
			name:        "invalid json returns INVALID_REQUEST_BODY",
			body:        `{invalid`,
			contentType: "application/json",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name: "unknown field returns INVALID_REQUEST_BODY",
			body: `{
				"plugin_type": "p",
				"name": "n",
				"config": {"a": "b"},
				"extra": 1
			}`,
			contentType: "application/json",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST_BODY",
		},
		{
			name: "missing plugin_type returns VALIDATION_FAILED",
			body: `{
				"name": "n",
				"config": {"a": "b"}
			}`,
			contentType: "application/json",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name: "missing config returns VALIDATION_FAILED",
			body: `{
				"plugin_type": "p",
				"name": "n"
			}`,
			contentType: "application/json",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name: "empty config returns VALIDATION_FAILED",
			body: `{
				"plugin_type": "p",
				"name": "n",
				"config": {}
			}`,
			contentType: "application/json",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name: "missing name returns VALIDATION_FAILED",
			body: `{
				"plugin_type": "p",
				"config": {"a": "b"}
			}`,
			contentType: "application/json",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_FAILED",
		},
		{
			name: "missing content type returns INVALID_CONTENT_TYPE",
			body: `{
				"plugin_type": "p",
				"name": "n",
				"config": {"a": "b"}
			}`,
			contentType: "",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name: "wrong content type returns INVALID_CONTENT_TYPE",
			body: `{
				"plugin_type": "p",
				"name": "n",
				"config": {"a": "b"}
			}`,
			contentType: "text/plain",
			setupMock:   func(_ *MockConnectorsService) {},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_CONTENT_TYPE",
		},
		{
			name: "invalid report returns 400 with details",
			body: `{
				"plugin_type": "p",
				"name": "n",
				"config": {"a": "b"}
			}`,
			contentType: "application/json",
			setupMock: func(m *MockConnectorsService) {
				m.EXPECT().
					ValidateConnector(mock.Anything, "p", mock.Anything).
					Return(invalidReport, nil).
					Once()
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "CONNECTOR_CONFIG_INVALID",
			wantDetails: invalidBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := assert.New(t)
			must := require.New(t)

			connectorsService := NewMockConnectorsService(t)
			tt.setupMock(connectorsService)

			handler := newTestHandler(connectorsService)

			rec := httptest.NewRecorder()
			rw := core_http_response.NewResponseWriter(rec)

			handler.ValidateConnector(rw, newValidateRequest(t, tt.body, tt.contentType))

			must.Equal(tt.wantStatus, rec.Code)

			if tt.wantCode != "" {
				var body errorResponseBody

				must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				is.Equal(tt.wantCode, body.Code)

				if len(body.Details) > 0 {
					var details transport.ValidateResponse

					must.NoError(json.Unmarshal(body.Details, &details))
					is.Equal(tt.wantDetails, details)
				}

				return
			}

			var body transport.ValidateResponse

			must.NoError(json.Unmarshal(rec.Body.Bytes(), &body))
			is.Equal(tt.wantBody, body)
		})
	}
}

func TestConnectorsHTTPHandler_ValidateConnector_InvalidFullBody(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	invalidReport := domain.ValidationReport{
		Valid: false,
		Steps: []domain.ValidationStep{{
			ID:     domain.StepCDC,
			Status: domain.CheckSeverityError,
			Checks: []domain.ValidationCheck{{
				ID:       "postgres.cdc.wal_level",
				Severity: domain.CheckSeverityError,
				Message:  "wal_level = 'replica'",
			}},
		}},
	}

	wantDetails := transport.ValidateResponse{
		Valid: false,
		Steps: []transport.ValidateStepResponse{{
			ID:     "cdc",
			Status: "error",
			Checks: []transport.ValidateCheckResponse{{
				ID:       "postgres.cdc.wal_level",
				Severity: "error",
				Message:  "wal_level = 'replica'",
			}},
		}},
	}

	type fullErrorBody struct {
		Code    string                     `json:"code"`
		Message string                     `json:"message"`
		Details transport.ValidateResponse `json:"details"`
	}

	wantBody := fullErrorBody{
		Code:    "CONNECTOR_CONFIG_INVALID",
		Message: "Connector configuration is invalid",
		Details: wantDetails,
	}

	connectorsService := NewMockConnectorsService(t)
	connectorsService.EXPECT().
		ValidateConnector(mock.Anything, "p", mock.Anything).
		Return(invalidReport, nil).
		Once()

	handler := newTestHandler(connectorsService)

	body := `{
		"plugin_type": "p",
		"name": "n",
		"config": {"a": "b"}
	}`

	rec := httptest.NewRecorder()
	rw := core_http_response.NewResponseWriter(rec)

	handler.ValidateConnector(rw, newValidateRequest(t, body, "application/json"))

	must.Equal(http.StatusBadRequest, rec.Code)
	must.Equal("application/json", rec.Header().Get("Content-Type"))

	var got fullErrorBody

	must.NoError(json.Unmarshal(rec.Body.Bytes(), &got))
	must.Equal(wantBody, got)
}
