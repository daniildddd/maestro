package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daniildddd/maestro/internal/core/domain"
)

func TestNewAuditLogFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		page       int
		limit      int
		action     string
		actor      string
		wantPage   int
		wantLimit  int
		wantAction domain.Action
		wantActor  string
		wantErr    error
	}{
		{
			name:      "zero page and limit get defaults",
			page:      0,
			limit:     0,
			wantPage:  1,
			wantLimit: 20,
		},
		{
			name:       "explicit page limit and action kept",
			page:       3,
			limit:      50,
			action:     "connector.created",
			actor:      "admin",
			wantPage:   3,
			wantLimit:  50,
			wantAction: domain.ActionConnectorCreated,
			wantActor:  "admin",
		},
		{
			name:      "negative page normalized",
			page:      -5,
			limit:     10,
			wantPage:  1,
			wantLimit: 10,
		},
		{
			name:      "negative limit normalized",
			page:      1,
			limit:     -5,
			wantPage:  1,
			wantLimit: 20,
		},
		{
			name:      "limit capped at 100",
			page:      1,
			limit:     500,
			wantPage:  1,
			wantLimit: 100,
		},
		{
			name:    "unknown action rejected",
			page:    1,
			limit:   20,
			action:  "bag.action",
			wantErr: domain.ErrInvalidAuditAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			must := require.New(t)

			filter, err := domain.NewAuditLogFilter(tt.page, tt.limit, tt.action, tt.actor)

			if tt.wantErr != nil {
				must.ErrorIs(err, tt.wantErr)
				must.Nil(filter)

				return
			}

			must.NoError(err)
			must.Equal(tt.wantPage, filter.Page)
			must.Equal(tt.wantLimit, filter.Limit)
			must.Equal(tt.wantAction, filter.Action)
			must.Equal(tt.wantActor, filter.Actor)
		})
	}
}
