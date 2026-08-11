package hasher_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/security/hasher"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env returns config with parsed cost", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		unsetAllHasherEnv(t)
		t.Setenv("HASHER_COST", "14")

		cfg := hasher.NewConfigMust()

		must.NotEmpty(cfg)
		is.Equal(14, cfg.Cost)
	})

	t.Run("defaults applied when HASHER_COST unset", func(t *testing.T) {
		is := assert.New(t)

		unsetAllHasherEnv(t)

		cfg := hasher.NewConfigMust()

		is.Equal(10, cfg.Cost)
	})

	t.Run("invalid HASHER_COST panics", func(t *testing.T) {
		must := require.New(t)

		unsetAllHasherEnv(t)
		t.Setenv("HASHER_COST", "abc")

		var panicVal any

		func() {
			defer func() {
				panicVal = recover()
			}()

			hasher.NewConfigMust()
		}()

		must.NotNil(panicVal)

		err, ok := panicVal.(error)
		must.True(ok)
		must.Error(err)
	})
}

func unsetAllHasherEnv(t *testing.T) {
	t.Helper()

	keys := []string{"HASHER_COST"}
	for _, key := range keys {
		old, ok := os.LookupEnv(key)
		_ = os.Unsetenv(key) //nolint:errcheck // env var removal cannot fail in practice

		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(key, old) //nolint:errcheck,usetesting // cannot use t.Setenv inside t.Cleanup; restore original
			}
		})
	}
}
