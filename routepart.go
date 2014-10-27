package router

type RoutePart struct {
	params   []string
	handler  Handler
	parts    map[string]*RoutePart
	prefixes []*Prefix
}

type Prefix struct {
	value   string
	handler Handler
}

func newRoutePart() *RoutePart {
	return &RoutePart{
		parts: make(map[string]*RoutePart),
	}
}
