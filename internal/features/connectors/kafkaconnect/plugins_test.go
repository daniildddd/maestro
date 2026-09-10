package kafkaconnect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestHTTPClient_GetConnectorPlugins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, _ *http.Request)
		wantIDs []string
		wantIs  error
	}{
		{
			name: "returns connector plugins",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, []map[string]string{
					{
						"class":   "io.debezium.connector.postgresql.PostgresConnector",
						"type":    "source",
						"version": "3.5.2",
					},
					{
						"class":   "io.debezium.connector.mysql.MySqlConnector",
						"type":    "source",
						"version": "3.5.2",
					},
					{
						"class":   "org.apache.kafka.connect.file.FileStreamSinkConnector",
						"type":    "sink",
						"version": "4.0.0",
					},
				})
			},
			wantIDs: []string{
				"io.debezium.connector.postgresql.PostgresConnector",
				"io.debezium.connector.mysql.MySqlConnector",
				"org.apache.kafka.connect.file.FileStreamSinkConnector",
			},
		},
		{
			name: "server error maps to kafka connect unavailable",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantIs: domain.ErrKafkaConnectUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			var gotPath string

			client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.RequestURI()
				tt.handler(w, r)
			})

			plugins, err := client.GetConnectorPlugins(context.Background())

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal("/connector-plugins", gotPath)
			must.Len(plugins, len(tt.wantIDs))

			for i, id := range tt.wantIDs {
				must.Equal(id, plugins[i].ID)
			}
		})
	}
}

func TestHTTPClient_GetSMTPlugins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		handler    func(w http.ResponseWriter, _ *http.Request)
		wantIDs    []string
		wantPath   string
		wantIs     error
		checkQuery bool
	}{
		{
			name: "returns only transformation plugins",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, []map[string]any{
					{"class": "io.debezium.transforms.ExtractNewRecordState", "type": "transformation"},
					{"class": "io.debezium.connector.postgresql.PostgresConnector", "type": "sink"},
					{"class": "io.debezium.transforms.TimestampConverter", "type": "transformation"},
				})
			},
			wantIDs: []string{
				"io.debezium.transforms.ExtractNewRecordState",
				"io.debezium.transforms.TimestampConverter",
			},
			wantPath:   "/connector-plugins?connectorsOnly=false",
			checkQuery: true,
		},
		{
			name: "server error maps to kafka connect unavailable",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantIs: domain.ErrKafkaConnectUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			var gotPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.RequestURI()
				tt.handler(w, r)
			}))
			t.Cleanup(srv.Close)

			client := newTestClientOnServer(t, srv)

			plugins, err := client.GetSMTPlugins(context.Background())

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Len(plugins, len(tt.wantIDs))

			for i, id := range tt.wantIDs {
				must.Equal(id, plugins[i].ID)
			}

			if tt.checkQuery {
				must.Equal(tt.wantPath, gotPath)
			}
		})
	}
}

func TestHTTPClient_GetConnectorPluginSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pluginID string
		handler  http.HandlerFunc
		want     domain.ConnectorPluginSchema
		wantPath string
		wantIs   error
	}{
		{
			name:     "maps config keys to plugin fields",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, []map[string]any{
					{
						"name":          "database.hostname",
						"type":          "STRING",
						"required":      true,
						"default_value": nil,
						"importance":    "HIGH",
						"documentation": "IP address or hostname of the PostgreSQL database server.",
						"display_name":  "Database Hostname",
					},
					{
						"name":          "database.port",
						"type":          "INT",
						"required":      true,
						"default_value": "5432",
						"importance":    "HIGH",
						"documentation": nil,
						"display_name":  nil,
					},
				})
			},
			want: domain.ConnectorPluginSchema{
				Fields: []domain.ConnectorPluginField{
					{
						Name:        "database.hostname",
						Label:       strPtr("Database Hostname"),
						Description: strPtr("IP address or hostname of the PostgreSQL database server."),
						Type:        "STRING",
						Importance:  "HIGH",
						Required:    true,
					},
					{
						Name:       "database.port",
						Type:       "INT",
						Importance: "HIGH",
						Required:   true,
						Default:    strPtr("5432"),
					},
				},
			},
		},
		{
			name:     "smt plugin maps config keys to plugin fields",
			pluginID: "io.debezium.transforms.ExtractNewRecordState",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, []map[string]any{
					{
						"name":          "insert.field",
						"type":          "STRING",
						"required":      false,
						"default_value": "value",
						"importance":    "MEDIUM",
						"documentation": "Field to insert",
						"display_name":  "Insert field",
					},
				})
			},
			want: domain.ConnectorPluginSchema{Fields: []domain.ConnectorPluginField{{
				Name:        "insert.field",
				Label:       strPtr("Insert field"),
				Description: strPtr("Field to insert"),
				Type:        "STRING",
				Importance:  "MEDIUM",
				Required:    false,
				Default:     strPtr("value"),
			}}},
			wantPath: "/connector-plugins/io.debezium.transforms.ExtractNewRecordState/config",
		},
		{
			name:     "unknown plugin maps to connector plugin not found",
			pluginID: "unknown.Connector",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				writeJSON(t, w, map[string]string{"error_code": "404", "message": "Connector plugin not found"})
			},
			wantIs: domain.ErrConnectorPluginNotFound,
		},
		{
			name:     "server error maps to kafka connect unavailable",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantIs: domain.ErrKafkaConnectUnavailable,
		},
		{
			name:     "requests plugin config path",
			pluginID: "io.debezium.connector.postgresql.PostgresConnector",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, []map[string]string{})
			},
			want: domain.ConnectorPluginSchema{Fields: []domain.ConnectorPluginField{}},
			wantPath: "/connector-plugins/" +
				"io.debezium.connector.postgresql.PostgresConnector/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			var gotPath string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.RequestURI()
				tt.handler(w, r)
			}))
			t.Cleanup(srv.Close)

			client := newTestClientOnServer(t, srv)

			got, err := client.GetConnectorPluginSchema(context.Background(), tt.pluginID)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, got)

			if tt.wantPath != "" {
				must.Equal(tt.wantPath, gotPath)
			}
		})
	}
}

func TestHTTPClient_GetConnectorPluginSchemaWithValues(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(t, w, []map[string]any{
				{
					"name":       "insert.field",
					"type":       "STRING",
					"required":   false,
					"importance": "MEDIUM",
				},
			})
		case http.MethodPut:
			writeJSON(t, w, map[string]any{
				"configs": []map[string]any{
					{
						"value": map[string]any{
							"name":               "insert.field",
							"recommended_values": []string{"op", "ts_ms", "op", "ts_ms"},
						},
					},
					{
						"value": map[string]any{
							"name":               "database.hostname",
							"recommended_values": []any{},
						},
					},
				},
			})
		default:
			http.NotFound(w, nil)
		}
	}))
	t.Cleanup(srv.Close)

	client := newTestClientOnServer(t, srv)

	schema, err := client.GetConnectorPluginSchemaWithValues(
		context.Background(),
		"io.debezium.transforms.ExtractNewRecordState",
	)

	must.NoError(err)
	must.Equal(
		domain.ConnectorPluginSchema{Fields: []domain.ConnectorPluginField{{
			Name:       "insert.field",
			Type:       "STRING",
			Required:   false,
			Importance: "MEDIUM",
			Values:     []string{"op", "ts_ms"},
		}}},
		schema,
	)
}

func TestHTTPClient_ValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, _ *http.Request)
		want    []domain.ValidationCheck
		wantIs  error
	}{
		{
			name: "valid config returns ok check",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, map[string]any{
					"configs": []map[string]any{
						{"value": map[string]any{"name": "database.hostname", "errors": []any{}}},
						{"value": map[string]any{"name": "database.port", "errors": []any{}}},
					},
				})
			},
			want: []domain.ValidationCheck{{
				ID:       "config.schema",
				Message:  "configuration accepted by Kafka Connect",
				Severity: domain.CheckSeverityOK,
			}},
		},
		{
			name: "invalid fields return error checks",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, map[string]any{
					"configs": []map[string]any{
						{
							"value": map[string]any{
								"name":   "database.hostname",
								"errors": []string{"hostname is required"},
							},
						},
						{
							"value": map[string]any{
								"name":   "database.port",
								"errors": []string{"", "not a number"},
							},
						},
					},
				})
			},
			want: []domain.ValidationCheck{
				{
					ID:       "config.database.hostname",
					Message:  "hostname is required",
					Severity: domain.CheckSeverityError,
					Field:    "database.hostname",
				},
				{
					ID:       "config.database.port",
					Message:  "; not a number",
					Severity: domain.CheckSeverityError,
					Field:    "database.port",
				},
			},
		},
		{
			name: "missing plugin maps to connector plugin not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, nil)
			},
			wantIs: domain.ErrConnectorPluginNotFound,
		},
		{
			name: "server error maps to kafka connect unavailable",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantIs: domain.ErrKafkaConnectUnavailable,
		},
		{
			name: "multiple errors join into one message",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, map[string]any{
					"configs": []map[string]any{
						{
							"value": map[string]any{
								"name":   "database.port",
								"errors": []string{"not a number", "must be positive"},
							},
						},
					},
				})
			},
			want: []domain.ValidationCheck{{
				ID:       "config.database.port",
				Message:  "not a number; must be positive",
				Severity: domain.CheckSeverityError,
				Field:    "database.port",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			srv := httptest.NewServer(http.HandlerFunc(tt.handler))
			t.Cleanup(srv.Close)

			client := newTestClientOnServer(t, srv)

			checks, err := client.ValidateConfig(
				context.Background(),
				"io.debezium.connector.postgresql.PostgresConnector",
				map[string]string{"database.hostname": "pg-1"},
			)

			if tt.wantIs != nil {
				must.ErrorIs(err, tt.wantIs)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, checks)
		})
	}
}
