package cleanup_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/features/auth/cleanup"
)

//nolint:paralleltest // t.Setenv is incompatible with t.Parallel (affects process-wide env)
func TestNewConfigMust(t *testing.T) {
	t.Run("valid env does not panic and returns config", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		t.Setenv("AUTH_CLEANUP_INTERVAL", "45m")

		var (
			panicVal any
			config   cleanup.Config
		)

		func() {
			defer func() {
				panicVal = recover()
			}()

			config = cleanup.NewConfigMust()
		}()

		must.Nil(panicVal)
		is.Equal(45*time.Minute, config.Interval)
	})

	t.Run("invalid env panics with config error", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		t.Setenv("AUTH_CLEANUP_INTERVAL", "notaduration")

		var panicVal any

		func() {
			defer func() {
				panicVal = recover()
			}()

			cleanup.NewConfigMust()
		}()

		must.NotNil(panicVal)

		panicErr, ok := panicVal.(error)
		must.True(ok)

		is.Contains(panicErr.Error(), "get cleanup config")
	})
}
