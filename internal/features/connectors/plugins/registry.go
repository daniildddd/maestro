package plugins

type Registry struct {
	adapters []Adapter
}

func NewRegistry(adapters ...Adapter) *Registry {
	return &Registry{adapters: adapters}
}

func (r *Registry) For(class string) Adapter {
	for _, adapter := range r.adapters {
		if adapter.Match(class) {
			return adapter
		}
	}

	return nil
}
