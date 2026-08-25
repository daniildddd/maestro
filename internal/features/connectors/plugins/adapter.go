package plugins

import "github.com/daniildddd/maestro/internal/core/domain"

type Adapter interface {
	Match(class string) bool

	Canonicalize(config map[string]string) domain.SourceConfig
}
