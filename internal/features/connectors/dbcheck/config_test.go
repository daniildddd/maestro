package dbcheck_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/connectors/dbcheck"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed values", func(t *testing.T) {
		is := assert.New(t)

		unsetAllDbcheckEnv(t)
		t.Setenv("VALIDATE_DB_STEP_TIMEOUT", "7s")
		t.Setenv("VALIDATE_DB_MAX_TABLE_CHECKS", "11")

		cfg := dbcheck.NewConfigMust()

		is.Equal(7*time.Second, cfg.StepTimeout)
		is.Equal(11, cfg.MaxTableChecks)
	})

	t.Run("defaults applied when env unset", func(t *testing.T) {
		is := assert.New(t)

		unsetAllDbcheckEnv(t)

		cfg := dbcheck.NewConfigMust()

		is.Equal(10*time.Second, cfg.StepTimeout)
		is.Equal(50, cfg.MaxTableChecks)
	})

	t.Run("invalid duration panics", func(t *testing.T) {
		unsetAllDbcheckEnv(t)
		t.Setenv("VALIDATE_DB_STEP_TIMEOUT", "not-a-duration")

		mustPanic(t, func() { dbcheck.NewConfigMust() })
	})

	t.Run("invalid int panics", func(t *testing.T) {
		unsetAllDbcheckEnv(t)
		t.Setenv("VALIDATE_DB_MAX_TABLE_CHECKS", "not-a-number")

		mustPanic(t, func() { dbcheck.NewConfigMust() })
	})
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()

	var panicVal any

	func() {
		defer func() { panicVal = recover() }()

		fn()
	}()

	must := require.New(t)
	must.NotNil(panicVal)
	err, ok := panicVal.(error)
	must.True(ok)
	must.Error(err)
}

func unsetAllDbcheckEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"VALIDATE_DB_STEP_TIMEOUT", "VALIDATE_DB_MAX_TABLE_CHECKS",
	}
	for _, key := range keys {
		old, ok := os.LookupEnv(key)
		_ = os.Unsetenv(key) //nolint:errcheck // env var removal cannot fail in practice

		t.Cleanup(func() {
			if ok {
				//nolint:errcheck,usetesting // cannot use t.Setenv inside t.Cleanup; restore original
				_ = os.Setenv(key, old)
			}
		})
	}
}
