package router

import (
	"regexp"
)

type RoutePart struct {
	variables []string
	action    *Action
	glob      bool
	parts     map[string]*RoutePart
	params    []Param
	prefixes  []Prefix
}

func newRoutePart() *RoutePart {
	return &RoutePart{
		parts: make(map[string]*RoutePart),
	}
}

type Param struct {
	constraint *regexp.Regexp
	route      *RoutePart
	suffix     string
}

func newParam(constraint *regexp.Regexp, route *RoutePart, suffix string) Param {
	return Param{
		route:      route,
		suffix:     suffix,
		constraint: constraint,
	}
}

type Prefix struct {
	value  string
	action *Action
}
