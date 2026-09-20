package plugins_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/features/connectors/plugins"
)

func TestMongoDBAdapter_Match(t *testing.T) {
	t.Parallel()

	must := require.New(t)

	adapter := plugins.MongoDBAdapter{}

	must.True(adapter.Match("io.debezium.connector.mongodb.MongoDbConnector"))
	must.False(adapter.Match("io.debezium.connector.mysql.MySqlConnector"))
	must.False(adapter.Match(""))
}

func TestMongoDBAdapter_Canonicalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
		want   domain.SourceConfig
	}{
		{
			name: "replica set shows full seed list, not first host",
			config: map[string]string{
				"mongodb.connection.string": "mongodb://debezium:secret@mongo1:27017," +
					"mongo2:27017/?replicaSet=rs0",
			},
			want: domain.SourceConfig{
				Hostname: "mongo1:27017, mongo2:27017",
				User:     "debezium",
			},
		},
		{
			name: "srv connection string without port",
			config: map[string]string{
				"mongodb.connection.string": "mongodb+srv://mongo.example.com/?retrywrites=true",
				"mongodb.user":              "debezium",
			},
			want: domain.SourceConfig{
				Hostname: "mongo.example.com",
				User:     "debezium",
			},
		},
		{
			name: "connection string without credentials falls back to mongodb.user",
			config: map[string]string{
				"mongodb.connection.string": "mongodb://mongo:27017",
				"mongodb.user":              "debezium",
			},
			want: domain.SourceConfig{
				Hostname: "mongo",
				Port:     "27017",
				User:     "debezium",
			},
		},
		{
			name: "hosts property removed in Debezium 3.x fields empty hostport",
			config: map[string]string{
				"mongodb.hosts": "mongo1:27017,mongo2:27017",
				"mongodb.user":  "debezium",
			},
			want: domain.SourceConfig{
				User: "debezium",
			},
		},
		{
			name: "invalid connection string falls back to user",
			config: map[string]string{
				"mongodb.connection.string": "://bad-url",
				"mongodb.user":              "debezium",
			},
			want: domain.SourceConfig{
				User: "debezium",
			},
		},
		{
			name: "connection string without host falls back to user",
			config: map[string]string{
				"mongodb.connection.string": "/just/a/path",
				"mongodb.user":              "debezium",
			},
			want: domain.SourceConfig{
				User: "debezium",
			},
		},
		{
			name:   "empty config yields zero value",
			config: map[string]string{},
			want:   domain.SourceConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			got := plugins.MongoDBAdapter{}.Canonicalize(tt.config)

			must.Equal(tt.want, got)
		})
	}
}
