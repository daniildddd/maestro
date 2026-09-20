package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestNewConnectorPluginSchemaFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		importance string
		want       domain.ConnectorPluginSchemaFilter
		wantErr    bool
	}{
		{
			name:       "high importance accepted",
			importance: "high",
			want:       domain.ConnectorPluginSchemaFilter{Importance: domain.PluginSchemaFilterHigh},
		},
		{
			name:       "medium importance accepted",
			importance: "medium",
			want:       domain.ConnectorPluginSchemaFilter{Importance: domain.PluginSchemaFilterMedium},
		},
		{
			name:       "all importance accepted",
			importance: "all",
			want:       domain.ConnectorPluginSchemaFilter{Importance: domain.PluginSchemaFilterAll},
		},
		{
			name:       "empty defaults to high",
			importance: "",
			want:       domain.ConnectorPluginSchemaFilter{Importance: domain.PluginSchemaFilterHigh},
		},
		{
			name:       "unknown value rejected",
			importance: "critical",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			filter, err := domain.NewConnectorPluginSchemaFilter(tt.importance)

			if tt.wantErr {
				must.Error(err)
				must.Nil(filter)

				return
			}

			must.NoError(err)
			must.Equal(tt.want, *filter)
		})
	}
}

func TestConnectorPluginSchemaFilter_Apply(t *testing.T) {
	t.Parallel()

	fields := []domain.ConnectorPluginField{
		{
			Name:        "database.hostname",
			Label:       new("Hostname"),
			Description: new("Resolvable hostname of the database server."),
			Type:        "STRING",
			Importance:  domain.PluginImportanceHigh,
			Required:    true,
			Default:     new("localhost"),
			Values:      nil,
		},
		{
			Name:        "database.port",
			Label:       new("Port"),
			Description: new("Port of the database server."),
			Type:        "INT",
			Importance:  domain.PluginImportanceHigh,
			Required:    false,
			Default:     new("5432"),
			Values:      []string{"5432", "5433"},
		},
		{
			Name:        "plugin.name",
			Label:       new("Plugin"),
			Description: new("Logical decoding plugin installed on the server."),
			Type:        "STRING",
			Importance:  domain.PluginImportanceMedium,
			Required:    true,
			Default:     new("pgoutput"),
			Values:      []string{"pgoutput", "decoderbufs"},
		},
		{
			Name:        "slot.name",
			Label:       new("Slot"),
			Description: new("Logical decoding slot name."),
			Type:        "STRING",
			Importance:  domain.PluginImportanceMedium,
			Required:    false,
			Default:     nil,
			Values:      nil,
		},
		{
			Name:        "internal.clock",
			Label:       nil,
			Description: nil,
			Type:        "BOOLEAN",
			Importance:  domain.PluginImportanceLow,
			Required:    false,
			Default:     new("false"),
			Values:      nil,
		},
		{
			Name:        "obscure.required",
			Label:       new("Obscure"),
			Description: nil,
			Type:        "STRING",
			Importance:  domain.PluginImportanceLow,
			Required:    true,
			Default:     nil,
			Values:      nil,
		},
	}

	tests := []struct {
		name       string
		importance string
		want       []domain.ConnectorPluginField
	}{
		{
			name:       "high keeps high and required fields",
			importance: "high",
			want: []domain.ConnectorPluginField{
				{
					Name:        "database.hostname",
					Label:       new("Hostname"),
					Description: new("Resolvable hostname of the database server."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceHigh,
					Required:    true,
					Default:     new("localhost"),
					Values:      nil,
				},
				{
					Name:        "database.port",
					Label:       new("Port"),
					Description: new("Port of the database server."),
					Type:        "INT",
					Importance:  domain.PluginImportanceHigh,
					Required:    false,
					Default:     new("5432"),
					Values:      []string{"5432", "5433"},
				},
				{
					Name:        "plugin.name",
					Label:       new("Plugin"),
					Description: new("Logical decoding plugin installed on the server."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceMedium,
					Required:    true,
					Default:     new("pgoutput"),
					Values:      []string{"pgoutput", "decoderbufs"},
				},
				{
					Name:        "obscure.required",
					Label:       new("Obscure"),
					Description: nil,
					Type:        "STRING",
					Importance:  domain.PluginImportanceLow,
					Required:    true,
					Default:     nil,
					Values:      nil,
				},
			},
		},
		{
			name:       "medium adds medium fields",
			importance: "medium",
			want: []domain.ConnectorPluginField{
				{
					Name:        "database.hostname",
					Label:       new("Hostname"),
					Description: new("Resolvable hostname of the database server."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceHigh,
					Required:    true,
					Default:     new("localhost"),
					Values:      nil,
				},
				{
					Name:        "database.port",
					Label:       new("Port"),
					Description: new("Port of the database server."),
					Type:        "INT",
					Importance:  domain.PluginImportanceHigh,
					Required:    false,
					Default:     new("5432"),
					Values:      []string{"5432", "5433"},
				},
				{
					Name:        "plugin.name",
					Label:       new("Plugin"),
					Description: new("Logical decoding plugin installed on the server."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceMedium,
					Required:    true,
					Default:     new("pgoutput"),
					Values:      []string{"pgoutput", "decoderbufs"},
				},
				{
					Name:        "slot.name",
					Label:       new("Slot"),
					Description: new("Logical decoding slot name."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceMedium,
					Required:    false,
					Default:     nil,
					Values:      nil,
				},
				{
					Name:        "obscure.required",
					Label:       new("Obscure"),
					Description: nil,
					Type:        "STRING",
					Importance:  domain.PluginImportanceLow,
					Required:    true,
					Default:     nil,
					Values:      nil,
				},
			},
		},
		{
			name:       "all returns everything",
			importance: "all",
			want: []domain.ConnectorPluginField{
				{
					Name:        "database.hostname",
					Label:       new("Hostname"),
					Description: new("Resolvable hostname of the database server."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceHigh,
					Required:    true,
					Default:     new("localhost"),
					Values:      nil,
				},
				{
					Name:        "database.port",
					Label:       new("Port"),
					Description: new("Port of the database server."),
					Type:        "INT",
					Importance:  domain.PluginImportanceHigh,
					Required:    false,
					Default:     new("5432"),
					Values:      []string{"5432", "5433"},
				},
				{
					Name:        "plugin.name",
					Label:       new("Plugin"),
					Description: new("Logical decoding plugin installed on the server."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceMedium,
					Required:    true,
					Default:     new("pgoutput"),
					Values:      []string{"pgoutput", "decoderbufs"},
				},
				{
					Name:        "slot.name",
					Label:       new("Slot"),
					Description: new("Logical decoding slot name."),
					Type:        "STRING",
					Importance:  domain.PluginImportanceMedium,
					Required:    false,
					Default:     nil,
					Values:      nil,
				},
				{
					Name:        "internal.clock",
					Label:       nil,
					Description: nil,
					Type:        "BOOLEAN",
					Importance:  domain.PluginImportanceLow,
					Required:    false,
					Default:     new("false"),
					Values:      nil,
				},
				{
					Name:        "obscure.required",
					Label:       new("Obscure"),
					Description: nil,
					Type:        "STRING",
					Importance:  domain.PluginImportanceLow,
					Required:    true,
					Default:     nil,
					Values:      nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			filter, err := domain.NewConnectorPluginSchemaFilter(tt.importance)
			must.NoError(err)

			got := filter.Apply(fields)

			must.Equal(tt.want, got)
		})
	}
}
