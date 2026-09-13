//go:build integration

package dbcheck_test

import (
	"context"
	"fmt"
	"maps"
	"net"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/daniildddd/maestro/internal/core/domain"
	"github.com/daniildddd/maestro/internal/core/repository/postgres/pgxadapter"
	"github.com/daniildddd/maestro/internal/features/connectors/dbcheck"
)

const (
	testDBUser     = "maestro"
	testDBPassword = "maestro"
	testDBName     = "maestro"
)

type testDatabase struct {
	host string
	port string
	dsn  string
}

func newLogicalDatabase(t *testing.T) testDatabase {
	t.Helper()

	return newTestDatabase(t, "postgres", "-c", "wal_level=logical")
}

func newTestDatabase(t *testing.T, cmd ...string) testDatabase {
	t.Helper()

	ctx := context.Background()
	must := require.New(t)

	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForSQL("5432/tcp", "pgx", func(host string, port network.Port) string {
				return fmt.Sprintf(
					"postgres://%s:%s@%s/%s?sslmode=disable",
					testDBUser,
					testDBPassword,
					net.JoinHostPort(host, port.Port()),
					testDBName,
				)
			}).WithStartupTimeout(60 * time.Second),
		),
	}

	if len(cmd) > 0 {
		opts = append(opts, testcontainers.WithCmd(cmd...))
	}

	container, err := postgres.Run(ctx, "postgres:17-alpine", opts...)
	must.NoError(err)

	testcontainers.CleanupContainer(t, container)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	must.NoError(err)

	host, err := container.Host(ctx)
	must.NoError(err)

	mappedPort, err := container.MappedPort(ctx, "5432")
	must.NoError(err)

	return testDatabase{host: host, port: mappedPort.Port(), dsn: dsn}
}

func execSetup(t *testing.T, dsn string, statements ...string) {
	t.Helper()

	must := require.New(t)
	ctx := context.Background()

	pool, err := pgxadapter.NewPool(ctx, pgxadapter.Config{DSN: dsn, Timeout: 10 * time.Second})
	must.NoError(err)
	t.Cleanup(pool.Close)

	for _, stmt := range statements {
		_, err := pool.Exec(ctx, stmt)
		must.NoError(err)
	}
}

func checkerConfig(db testDatabase, extra map[string]string) map[string]string {
	config := map[string]string{
		"database.hostname": db.host,
		"database.port":     db.port,
		"database.user":     testDBUser,
		"database.password": testDBPassword,
		"database.dbname":   testDBName,
		"plugin.name":       "pgoutput",
	}

	maps.Copy(config, extra)

	return config
}

func newChecker() *dbcheck.PostgresChecker {
	return dbcheck.NewPostgresChecker(dbcheck.Config{
		StepTimeout:    10 * time.Second,
		MaxTableChecks: 50,
	})
}

func findCheck(
	steps []domain.ValidationStep,
	stepID domain.StepID,
	checkID string,
) (domain.ValidationCheck, bool) {
	for _, step := range steps {
		if step.ID != stepID {
			continue
		}

		for _, check := range step.Checks {
			if check.ID == checkID {
				return check, true
			}
		}
	}

	return domain.ValidationCheck{}, false
}

func assertNoErrors(t *testing.T, steps []domain.ValidationStep) {
	t.Helper()

	for _, step := range steps {
		for _, check := range step.Checks {
			assert.NotEqual(t, domain.CheckSeverityError, check.Severity,
				"unexpected error in step %q check %q: %s", step.ID, check.ID, check.Message)
		}
	}
}

func TestPostgresChecker_HealthySourceAllOK(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)
	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.orders (id SERIAL PRIMARY KEY, amount INT NOT NULL)",
		"CREATE PUBLICATION testpub FOR ALL TABLES",
	)

	steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
		"table.include.list": "public.orders",
		"publication.name":   "testpub",
	}))
	must.NoError(err)
	must.Len(steps, 4)

	assertNoErrors(t, steps)

	auth, found := findCheck(steps, domain.StepConnection, "postgres.connection.auth")
	must.True(found)
	is.Equal(domain.CheckSeverityOK, auth.Severity)

	walLevel, found := findCheck(steps, domain.StepCDC, "postgres.cdc.wal_level")
	must.True(found)
	is.Equal(domain.CheckSeverityOK, walLevel.Severity)

	readiness, found := findCheck(steps, domain.StepTables, "postgres.table.readiness")
	must.True(found)
	is.Equal(domain.CheckSeverityOK, readiness.Severity)
}

func TestPostgresChecker_WrongPassword(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)
	db := newLogicalDatabase(t)

	config := checkerConfig(db, nil)
	config["database.password"] = "wrong-password"

	steps, err := newChecker().Check(context.Background(), config)
	must.NoError(err)
	must.Len(steps, 4)
	is.Equal(domain.StepConnection, steps[0].ID)

	auth, found := findCheck(steps, domain.StepConnection, "postgres.connection.auth")
	must.True(found)
	is.Equal(domain.CheckSeverityError, auth.Severity)

	for _, step := range steps[1:] {
		for _, check := range step.Checks {
			is.Equal(domain.CheckSeveritySkipped, check.Severity)
		}
	}
}

func TestPostgresChecker_WalLevelNotLogical(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)
	db := newTestDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.orders (id SERIAL PRIMARY KEY)",
	)

	steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
		"table.include.list": "public.orders",
	}))
	must.NoError(err)

	walLevel, found := findCheck(steps, domain.StepCDC, "postgres.cdc.wal_level")
	must.True(found)
	is.Equal(domain.CheckSeverityError, walLevel.Severity)
}

func TestPostgresChecker_TableReadiness(t *testing.T) {
	t.Parallel()

	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.ready (id SERIAL PRIMARY KEY)",
		"CREATE TABLE public.nopk (id INT)",
		"CREATE UNLOGGED TABLE public.ephemeral (id SERIAL PRIMARY KEY, payload TEXT)",
		"CREATE PUBLICATION testpub FOR ALL TABLES",
	)

	tests := []struct {
		name         string
		table        string
		wantCheckID  string
		wantSeverity domain.CheckSeverity
	}{
		{
			name:         "missing table is an error",
			table:        "public.nope",
			wantCheckID:  "postgres.table.exists",
			wantSeverity: domain.CheckSeverityError,
		},
		{
			name:         "table without primary key warns",
			table:        "public.nopk",
			wantCheckID:  "postgres.table.primary_key",
			wantSeverity: domain.CheckSeverityWarning,
		},
		{
			name:         "unlogged table is an error",
			table:        "public.ephemeral",
			wantCheckID:  "postgres.table.unlogged",
			wantSeverity: domain.CheckSeverityError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)
			is := assert.New(t)

			steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
				"table.include.list": tt.table,
				"publication.name":   "testpub",
			}))
			must.NoError(err)

			found, ok := findCheck(steps, domain.StepTables, tt.wantCheckID)
			must.True(ok, "check %q should be present", tt.wantCheckID)
			is.Equal(tt.wantSeverity, found.Severity)
		})
	}
}

func TestPostgresChecker_UserWithoutReplication(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)
	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"CREATE ROLE lowuser LOGIN PASSWORD 'lowuser'",
		"GRANT CONNECT ON DATABASE maestro TO lowuser",
	)

	config := checkerConfig(db, nil)
	config["database.user"] = "lowuser"
	config["database.password"] = "lowuser"

	steps, err := newChecker().Check(context.Background(), config)
	must.NoError(err)

	replication, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.replication")
	must.True(found)
	is.Equal(domain.CheckSeverityError, replication.Severity)
	is.NotEmpty(replication.FixHint)
}

func TestPostgresChecker_MissingRequiredFields(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)

	steps, err := newChecker().Check(context.Background(), map[string]string{})
	must.NoError(err)
	must.NotEmpty(steps)
	is.Equal(domain.StepConnection, steps[0].ID)

	for _, check := range steps[0].Checks {
		is.Equal(domain.CheckSeverityError, check.Severity)
	}
}

func TestPostgresChecker_DeniedPrivileges(t *testing.T) {
	t.Parallel()

	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"CREATE ROLE lowuser LOGIN PASSWORD 'lowuser'",
		"GRANT CONNECT ON DATABASE maestro TO lowuser",
		"ALTER ROLE lowuser REPLICATION",
		"CREATE SCHEMA private",
		"CREATE TABLE private.orders (id SERIAL PRIMARY KEY)",
		"CREATE TABLE private.noselect (id SERIAL PRIMARY KEY)",
		"GRANT USAGE ON SCHEMA private TO lowuser",
		"CREATE SCHEMA secret",
		"CREATE TABLE secret.t (id SERIAL PRIMARY KEY)",
	)

	lowuser := func(extra map[string]string) map[string]string {
		config := checkerConfig(db, extra)
		config["database.user"] = "lowuser"
		config["database.password"] = "lowuser"

		return config
	}

	t.Run("select denied is an error", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)

		steps, err := newChecker().Check(context.Background(), lowuser(map[string]string{
			"table.include.list": "private.noselect",
		}))
		must.NoError(err)

		denied, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.select")
		must.True(found, "select check should be present")
		is.Equal(domain.CheckSeverityError, denied.Severity)
		is.Equal("private.noselect", denied.Table)
		is.NotEmpty(denied.FixHint)
	})

	t.Run("usage denied is an error", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)

		steps, err := newChecker().Check(context.Background(), lowuser(map[string]string{
			"table.include.list": "secret.t",
		}))
		must.NoError(err)

		denied, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.schema_usage")
		must.True(found, "schema usage check should be present")
		is.Equal(domain.CheckSeverityError, denied.Severity)
		is.Equal("table.include.list", denied.Field)
		is.NotEmpty(denied.FixHint)
	})

	t.Run("create denied warns", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)

		steps, err := newChecker().Check(context.Background(), lowuser(map[string]string{
			"table.include.list": "private.orders",
		}))
		must.NoError(err)

		create, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.create")
		must.True(found, "create check should be present")
		is.Equal(domain.CheckSeverityWarning, create.Severity)
		is.NotEmpty(create.FixHint)
	})

	t.Run("autocreate disabled skips create check", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)

		steps, err := newChecker().Check(context.Background(), lowuser(map[string]string{
			"table.include.list":          "private.orders",
			"publication.autocreate.mode": "disabled",
		}))
		must.NoError(err)

		_, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.create")
		must.False(found, "create check should be skipped when autocreate is disabled")
	})
}

func TestPostgresChecker_PublicationVariants(t *testing.T) {
	t.Parallel()

	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.a (id SERIAL PRIMARY KEY)",
		"CREATE TABLE public.b (id SERIAL PRIMARY KEY)",
		"CREATE PUBLICATION partial FOR TABLE public.a",
	)

	tests := []struct {
		name         string
		config       map[string]string
		stepID       domain.StepID
		wantCheckID  string
		wantSeverity domain.CheckSeverity
		wantFixHint  bool
	}{
		{
			name: "missing publication with autocreate disabled is an error",
			config: map[string]string{
				"publication.name":            "nosuchpub",
				"publication.autocreate.mode": "disabled",
			},
			stepID:       domain.StepCDC,
			wantCheckID:  "postgres.cdc.publication",
			wantSeverity: domain.CheckSeverityError,
			wantFixHint:  true,
		},
		{
			name: "missing publication with autocreate enabled warns",
			config: map[string]string{
				"publication.name": "nosuchpub",
			},
			stepID:       domain.StepCDC,
			wantCheckID:  "postgres.cdc.publication",
			wantSeverity: domain.CheckSeverityWarning,
		},
		{
			name: "publication without all tables warns",
			config: map[string]string{
				"publication.name": "partial",
			},
			stepID:       domain.StepCDC,
			wantCheckID:  "postgres.cdc.puballtables",
			wantSeverity: domain.CheckSeverityWarning,
		},
		{
			name: "table not in publication warns",
			config: map[string]string{
				"table.include.list": "public.b",
				"publication.name":   "partial",
			},
			stepID:       domain.StepTables,
			wantCheckID:  "postgres.table.publication",
			wantSeverity: domain.CheckSeverityWarning,
		},
		{
			name: "table not in publication errors when autocreate disabled",
			config: map[string]string{
				"table.include.list":          "public.b",
				"publication.name":            "partial",
				"publication.autocreate.mode": "disabled",
			},
			stepID:       domain.StepTables,
			wantCheckID:  "postgres.table.publication",
			wantSeverity: domain.CheckSeverityError,
			wantFixHint:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)
			is := assert.New(t)

			steps, err := newChecker().Check(context.Background(), checkerConfig(db, tt.config))
			must.NoError(err)

			got, found := findCheck(steps, tt.stepID, tt.wantCheckID)
			must.True(found, "check %q should be present", tt.wantCheckID)
			is.Equal(tt.wantSeverity, got.Severity)

			if tt.wantFixHint {
				is.NotEmpty(got.FixHint)
			}
		})
	}
}

func TestPostgresChecker_ReplicaIdentity(t *testing.T) {
	t.Parallel()

	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.has_pk (id SERIAL PRIMARY KEY)",
		"CREATE TABLE public.full_id (id INT NOT NULL)",
		"ALTER TABLE public.full_id REPLICA IDENTITY FULL",
		"CREATE TABLE public.no_id (id INT)",
		"ALTER TABLE public.no_id REPLICA IDENTITY NOTHING",
		"CREATE TABLE public.uniq_id (id INT NOT NULL)",
		"CREATE UNIQUE INDEX uniq_id_idx ON public.uniq_id (id)",
		"ALTER TABLE public.uniq_id REPLICA IDENTITY USING INDEX uniq_id_idx",
		"CREATE PUBLICATION testpub FOR ALL TABLES",
	)

	tests := []struct {
		name         string
		table        string
		wantSeverity domain.CheckSeverity
		wantMessage  string
	}{
		{
			name:         "table with primary key is ok",
			table:        "public.has_pk",
			wantSeverity: domain.CheckSeverityOK,
			wantMessage:  "has a primary key",
		},
		{
			name:         "replica identity full warns",
			table:        "public.full_id",
			wantSeverity: domain.CheckSeverityWarning,
			wantMessage:  "REPLICA IDENTITY FULL",
		},
		{
			name:         "replica identity nothing warns",
			table:        "public.no_id",
			wantSeverity: domain.CheckSeverityWarning,
			wantMessage:  "REPLICA IDENTITY NOTHING",
		},
		{
			name:         "replica identity using unique index is ok",
			table:        "public.uniq_id",
			wantSeverity: domain.CheckSeverityOK,
			wantMessage:  "unique index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)
			is := assert.New(t)

			steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
				"table.include.list": tt.table,
				"publication.name":   "testpub",
			}))
			must.NoError(err)

			got, found := findCheck(steps, domain.StepTables, "postgres.table.primary_key")
			must.True(found, "primary key check should be present")
			is.Equal(tt.wantSeverity, got.Severity)
			is.Contains(got.Message, tt.wantMessage)
		})
	}
}

func TestPostgresChecker_DecodingPluginMissing(t *testing.T) {
	t.Parallel()

	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
	)

	for _, plugin := range []string{"wal2json", "decoderbufs"} {
		t.Run(plugin+" not installed is an error", func(t *testing.T) {
			t.Parallel()

			must := require.New(t)
			is := assert.New(t)

			steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
				"plugin.name": plugin,
			}))
			must.NoError(err)

			got, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.plugin")
			must.True(found, "plugin check should be present")
			is.Equal(domain.CheckSeverityError, got.Severity)
			is.NotEmpty(got.FixHint)
		})
	}
}

func TestPostgresChecker_TableFilterEdgeCases(t *testing.T) {
	t.Parallel()

	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.orders (id SERIAL PRIMARY KEY)",
		"CREATE PUBLICATION testpub FOR ALL TABLES",
	)

	t.Run("regex filters are reported as skipped", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)

		steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
			"table.include.list": "public.orders*",
			"publication.name":   "testpub",
		}))
		must.NoError(err)

		got, found := findCheck(steps, domain.StepTables, "postgres.table.regex")
		must.True(found, "regex check should be present")
		is.Equal(domain.CheckSeveritySkipped, got.Severity)
	})

	t.Run("problems truncate at max table checks", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)

		checker := dbcheck.NewPostgresChecker(dbcheck.Config{
			StepTimeout:    10 * time.Second,
			MaxTableChecks: 1,
		})

		steps, err := checker.Check(context.Background(), checkerConfig(db, map[string]string{
			"table.include.list": "public.m1, public.m2",
			"publication.name":   "testpub",
		}))
		must.NoError(err)

		missing, found := findCheck(steps, domain.StepTables, "postgres.table.exists")
		must.True(found, "missing table error should be present")
		is.Equal(domain.CheckSeverityError, missing.Severity)

		truncated, found := findCheck(steps, domain.StepTables, "postgres.table.problems")
		must.True(found, "truncation warning should be present")
		is.Equal(domain.CheckSeverityWarning, truncated.Severity)
		is.Contains(truncated.Message, "truncated at 1")
	})
}

func TestPostgresChecker_PluginUnknown(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)
	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
	)

	steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
		"plugin.name": "mystery",
	}))
	must.NoError(err)

	got, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.plugin")
	must.True(found, "plugin check should be present")
	is.Equal(domain.CheckSeverityWarning, got.Severity)
	is.Equal("plugin.name", got.Field)
	is.Contains(got.Message, "unknown")
}

func TestPostgresChecker_SelectSkippedVariants(t *testing.T) {
	t.Parallel()

	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.orders (id SERIAL PRIMARY KEY)",
		"CREATE PUBLICATION testpub FOR ALL TABLES",
	)

	t.Run("empty include list skips select and warns capture", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)

		steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
			"table.include.list": "",
			"publication.name":   "testpub",
		}))
		must.NoError(err)

		sel, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.select")
		must.True(found, "select check should be present")
		is.Equal(domain.CheckSeveritySkipped, sel.Severity)

		capture, found := findCheck(steps, domain.StepTables, "postgres.table.capture")
		must.True(found, "capture check should be present")
		is.Equal(domain.CheckSeverityWarning, capture.Severity)
	})

	t.Run("all missing tables skip select", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)

		steps, err := newChecker().Check(context.Background(), checkerConfig(db, map[string]string{
			"table.include.list": "public.nosuch1, public.nosuch2",
			"publication.name":   "testpub",
		}))
		must.NoError(err)

		sel, found := findCheck(steps, domain.StepPermissions, "postgres.privilege.select")
		must.True(found, "select check should be present")
		is.Equal(domain.CheckSeveritySkipped, sel.Severity)
	})
}

func TestPostgresChecker_CdcLimits(t *testing.T) {
	t.Parallel()

	t.Run("default reports retention slots senders", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)
		db := newLogicalDatabase(t)

		execSetup(t, db.dsn, "ALTER ROLE maestro REPLICATION")

		steps, err := newChecker().Check(context.Background(), checkerConfig(db, nil))
		must.NoError(err)

		for _, id := range []string{
			"postgres.cdc.wal_retention",
			"postgres.cdc.replication_slots",
			"postgres.cdc.wal_senders",
		} {
			got, found := findCheck(steps, domain.StepCDC, id)
			must.True(found, "check %q should be present", id)
			is.NotEqual(domain.CheckSeveritySkipped, got.Severity)
			is.NotEmpty(got.Message)
		}
	})

	t.Run("unlimited retention is ok", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)
		db := newLogicalDatabase(t)

		execSetup(t, db.dsn,
			"ALTER ROLE maestro REPLICATION",
			"ALTER SYSTEM SET max_slot_wal_keep_size = '-1'",
			"SELECT pg_reload_conf()",
		)

		steps, err := newChecker().Check(context.Background(), checkerConfig(db, nil))
		must.NoError(err)

		got, found := findCheck(steps, domain.StepCDC, "postgres.cdc.wal_retention")
		must.True(found, "wal_retention check should be present")
		is.Equal(domain.CheckSeverityOK, got.Severity)
		is.Contains(got.Message, "unlimited")
	})

	t.Run("no free replication slots is an error", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)
		db := newTestDatabase(t, "postgres", "-c", "wal_level=logical", "-c", "max_replication_slots=1")

		execSetup(t, db.dsn,
			"ALTER ROLE maestro REPLICATION",
			"SELECT pg_create_logical_replication_slot('s1', 'pgoutput')",
		)

		steps, err := newChecker().Check(context.Background(), checkerConfig(db, nil))
		must.NoError(err)

		got, found := findCheck(steps, domain.StepCDC, "postgres.cdc.replication_slots")
		must.True(found, "replication_slots check should be present")
		is.Equal(domain.CheckSeverityError, got.Severity)
		is.NotEmpty(got.FixHint)
	})

	t.Run("no free wal senders is an error", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		is := assert.New(t)
		db := newTestDatabase(t, "postgres", "-c", "wal_level=logical", "-c", "max_wal_senders=0")

		execSetup(t, db.dsn, "ALTER ROLE maestro REPLICATION")

		steps, err := newChecker().Check(context.Background(), checkerConfig(db, nil))
		must.NoError(err)

		got, found := findCheck(steps, domain.StepCDC, "postgres.cdc.wal_senders")
		must.True(found, "wal_senders check should be present")
		is.Equal(domain.CheckSeverityError, got.Severity)
		is.NotEmpty(got.FixHint)
	})
}

func TestPostgresChecker_ConfigWarningsMerged(t *testing.T) {
	t.Parallel()

	must := require.New(t)
	is := assert.New(t)
	db := newLogicalDatabase(t)

	execSetup(t, db.dsn,
		"ALTER ROLE maestro REPLICATION",
		"CREATE TABLE public.orders (id SERIAL PRIMARY KEY)",
		"CREATE PUBLICATION testpub FOR ALL TABLES",
	)

	config := checkerConfig(db, map[string]string{
		"table.include.list": "public.orders",
		"publication.name":   "testpub",
	})
	delete(config, "database.dbname")

	steps, err := newChecker().Check(context.Background(), config)
	must.NoError(err)
	must.Len(steps, 4)

	warn, found := findCheck(steps, domain.StepConnection, "postgres.connection.config")
	must.True(found, "config warning should be present")
	is.Equal(domain.CheckSeverityWarning, warn.Severity)
	is.Equal("database.dbname", warn.Field)

	auth, found := findCheck(steps, domain.StepConnection, "postgres.connection.auth")
	must.True(found)
	is.Equal(domain.CheckSeverityOK, auth.Severity)

	assertNoErrors(t, steps)
}
