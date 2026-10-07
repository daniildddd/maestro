package kafkaconnect_test

import (
	"net/http"
	"testing"

	gotest "github.com/allure-framework/allure-go/commons/gotest" //nolint:revive // package clause is allure, gotest keeps call sites readable
	allurerequire "github.com/allure-framework/allure-go/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestAllureHTTPClient_GetConnectors(t *testing.T) { //nolint:paralleltest // allure gotest.Test owns subtest execution
	gotest.Test(t, "list connectors captures connect exchange", func(a *gotest.Context) {
		must := allurerequire.New(a)

		var connectors []domain.Connector

		var listErr error

		a.Step("serve fake connect expansion", func(a *gotest.Context) {
			srv := newAllureTestServer(a, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(a.T(), w, listResponse{
					"pg-connector": newExpansion(
						"io.debezium.connector.postgresql.PostgresConnector",
						"RUNNING",
						map[string]string{
							"database.hostname": "maestro-postgres",
							"database.port":     "5432",
						},
					),
				})
			}))

			client := newTestClientOnServer(a.T(), srv)
			connectors, listErr = client.GetConnectors(a.Context())
		})

		a.Step("assert mapped connectors", func(_ *gotest.Context) {
			must.NoError(listErr)
			must.Len(connectors, 1)
			must.Equal("pg-connector", connectors[0].Name)
		})
	},
		gotest.WithEpic("Connectors"),
		gotest.WithFeature("List connectors"),
		gotest.WithStory("Kafka Connect HTTP exchange"),
		gotest.WithSeverity("normal"),
		gotest.WithOwner("qa-intern"),
	)
}
