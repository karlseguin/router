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
}

func newParam(constraint *regexp.Regexp, route *RoutePart) Param {
	return Param{
		constraint: constraint,
		route:      route,
	}
}

type Prefix struct {
	value  string
	action *Action
}
