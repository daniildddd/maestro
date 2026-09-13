package dbcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestSplitList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{
			name:  "single entry returned as is",
			value: "public.orders",
			want:  []string{"public.orders"},
		},
		{
			name:  "spaces around commas trimmed",
			value: " public.orders , public.payments ",
			want:  []string{"public.orders", "public.payments"},
		},
		{
			name:  "only empty parts yields empty list",
			value: ",,",
			want:  []string{},
		},
		{
			name:  "empty value yields empty list",
			value: "",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, splitList(tt.value))
		})
	}
}

func TestSplitSchemaTable(t *testing.T) {
	t.Parallel()

	schema, table := splitSchemaTable("public.orders")
	assert.Equal(t, "public", schema)
	assert.Equal(t, "orders", table)

	schema, table = splitSchemaTable("orders")
	assert.Equal(t, "public", schema)
	assert.Equal(t, "orders", table)
}

func TestIsRegexFilter(t *testing.T) {
	t.Parallel()

	assert.False(t, isRegexFilter("public.orders"))
	assert.True(t, isRegexFilter("public.orders*"))
	assert.True(t, isRegexFilter("public.(orders|payments)"))
}

func TestCheckBuilders(t *testing.T) {
	t.Parallel()

	base := check("postgres.cdc.wal_level", domain.CheckSeverityError, "wal_level = %q", "replica")
	require.Equal(t, "postgres.cdc.wal_level", base.ID)
	require.Equal(t, domain.CheckSeverityError, base.Severity)
	require.Equal(t, `wal_level = "replica"`, base.Message)
	require.Empty(t, base.Field)
	require.Empty(t, base.Table)
	require.Empty(t, base.FixHint)

	withField := fieldCheck("postgres.privilege.connect", domain.CheckSeverityError,
		"database.user", "user %q has no CONNECT privilege", "debezium")
	require.Equal(t, "database.user", withField.Field)
	require.Contains(t, withField.Message, "debezium")

	withTable := tableCheck("postgres.table.exists", domain.CheckSeverityError,
		"public.orders", "table %q does not exist", "public.orders")
	require.Equal(t, "public.orders", withTable.Table)

	withFix := fixCheck(base, "ALTER SYSTEM SET wal_level = logical;")
	require.Equal(t, "ALTER SYSTEM SET wal_level = logical;", withFix.FixHint)
	require.Equal(t, base.ID, withFix.ID)
	require.Equal(t, base.Message, withFix.Message)
}

func TestSkippedStep(t *testing.T) {
	t.Parallel()

	step := skippedStep(domain.StepCDC, "connection failed")

	require.Equal(t, domain.StepCDC, step.ID)
	require.Len(t, step.Checks, 1)
	require.Equal(t, string(domain.StepCDC)+".skipped", step.Checks[0].ID)
	require.Equal(t, domain.CheckSeveritySkipped, step.Checks[0].Severity)
	require.Equal(t, "connection failed", step.Checks[0].Message)
}

func TestSkippedDBSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		after   domain.StepID
		wantIDs []domain.StepID
	}{
		{
			name:  "connection failure skips everything after",
			after: domain.StepConnection,
			wantIDs: []domain.StepID{
				domain.StepPermissions,
				domain.StepCDC,
				domain.StepTables,
			},
		},
		{
			name:  "permissions failure skips cdc and tables",
			after: domain.StepPermissions,
			wantIDs: []domain.StepID{
				domain.StepCDC,
				domain.StepTables,
			},
		},
		{
			name:  "cdc failure skips tables",
			after: domain.StepCDC,
			wantIDs: []domain.StepID{
				domain.StepTables,
			},
		},
		{
			name:    "tables failure skips nothing",
			after:   domain.StepTables,
			wantIDs: nil,
		},
		{
			name:    "unknown step skips nothing",
			after:   domain.StepID("bogus"),
			wantIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			steps := skippedDBSteps(tt.after, "boom")

			require.Len(t, steps, len(tt.wantIDs))

			for i, stepID := range tt.wantIDs {
				require.Equal(t, stepID, steps[i].ID)
				require.Len(t, steps[i].Checks, 1)
				require.Equal(t, "boom", steps[i].Checks[0].Message)
			}
		})
	}
}
