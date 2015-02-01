package router

type RoutePart struct {
	params   []Param
	action   *Action
	glob     bool
	parts    map[string]*RoutePart
	prefixes []Prefix
}

type Prefix struct {
	value  string
	action *Action
}

type Param struct {
	name string
	// constraint *regexp.Regexp
}

func newRoutePart() *RoutePart {
	return &RoutePart{
		parts: make(map[string]*RoutePart),
	}
}
