package domain

import (
	"errors"
	"fmt"
)

const (
	PluginImportanceHigh   = "HIGH"
	PluginImportanceMedium = "MEDIUM"
	PluginImportanceLow    = "LOW"
)

const (
	PluginSchemaFilterHigh   = "high"
	PluginSchemaFilterMedium = "medium"
	PluginSchemaFilterAll    = "all"
)

var (
	ErrConnectorPluginNotFound = errors.New("connector plugin not found")
	ErrSMTPluginNotFound       = errors.New("smt plugin not found")
)

type ConnectorPlugin struct {
	ID string
}

type ConnectorPluginField struct {
	Name        string
	Label       *string
	Description *string
	Type        string
	Importance  string
	Required    bool
	Default     *string
	Values      []string
}

type ConnectorPluginSchema struct {
	Fields []ConnectorPluginField
}

type ConnectorPluginSchemaFilter struct {
	Importance string
}

func NewConnectorPluginSchemaFilter(importance string) (*ConnectorPluginSchemaFilter, error) {
	const op = "core.domain.NewConnectorPluginSchemaFilter"

	f := &ConnectorPluginSchemaFilter{Importance: importance}
	f.Normalize()

	if err := f.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return f, nil
}

func (f *ConnectorPluginSchemaFilter) Normalize() {
	if f.Importance == "" {
		f.Importance = PluginSchemaFilterHigh
	}
}

func (f *ConnectorPluginSchemaFilter) Validate() error {
	switch f.Importance {
	case PluginSchemaFilterHigh, PluginSchemaFilterMedium, PluginSchemaFilterAll:
		return nil
	default:
		return errors.New("importance must be one of: high, medium, all")
	}
}

func (f *ConnectorPluginSchemaFilter) Apply(fields []ConnectorPluginField) []ConnectorPluginField {
	matched := make([]ConnectorPluginField, 0, len(fields))

	for _, field := range fields {
		if f.matches(field) {
			matched = append(matched, field)
		}
	}

	return matched
}

func (f *ConnectorPluginSchemaFilter) matches(field ConnectorPluginField) bool {
	if field.Required {
		return true
	}

	switch f.Importance {
	case PluginSchemaFilterAll:
		return true
	case PluginSchemaFilterMedium:
		return field.Importance == PluginImportanceHigh || field.Importance == PluginImportanceMedium
	default:
		return field.Importance == PluginImportanceHigh
	}
}
