package router

type RoutePart struct {
	params  []string
	handler Handler
	parts   map[string]*RoutePart
}

func newRoutePart() *RoutePart {
	return &RoutePart{
		parts: make(map[string]*RoutePart),
	}
}
