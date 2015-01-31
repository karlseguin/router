package router

type RoutePart struct {
	params   []string
	action   *Action
	glob     bool
	parts    map[string]*RoutePart
	prefixes []*Prefix
}

type Prefix struct {
	value  string
	action *Action
}

func newRoutePart() *RoutePart {
	return &RoutePart{
		parts: make(map[string]*RoutePart),
	}
}
