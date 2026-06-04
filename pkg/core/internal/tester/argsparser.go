package tester

import (
	"flag"
	"fmt"
)

type ArgsParser interface {
	DefineBool(value *bool, name string, defaultValue ...bool)
	DefineInt(value *int, name string, defaultValue ...int)
	DefineInt64(value *int64, name string, defaultValue ...int64)
	DefineString(value *string, name string, defaultValue ...string)
	Parse()
}

type argsParser struct {
	prefix string

	keys []string
	vals []any
}

func Args(prefix string) ArgsParser {
	return &argsParser{
		prefix: prefix,
		keys:   make([]string, 0),
		vals:   make([]any, 0),
	}
}

func (p *argsParser) DefineBool(value *bool, name string, defaultValue ...bool) {
	var dvalue bool
	if len(defaultValue) > 0 {
		dvalue = defaultValue[0]
	}

	flag.BoolVar(value, p.prefix+name, dvalue, "")
	p.keys = append(p.keys, name)
	p.vals = append(p.vals, value)
}

func (p *argsParser) DefineInt(value *int, name string, defaultValue ...int) {
	var dvalue int
	if len(defaultValue) > 0 {
		dvalue = defaultValue[0]
	}

	flag.IntVar(value, p.prefix+name, dvalue, "")
	p.keys = append(p.keys, name)
	p.vals = append(p.vals, value)
}

func (p *argsParser) DefineInt64(value *int64, name string, defaultValue ...int64) {
	var dvalue int64
	if len(defaultValue) > 0 {
		dvalue = defaultValue[0]
	}

	flag.Int64Var(value, p.prefix+name, dvalue, "")
	p.keys = append(p.keys, name)
	p.vals = append(p.vals, value)
}

func (p *argsParser) DefineString(value *string, name string, defaultValue ...string) {
	var dvalue string
	if len(defaultValue) > 0 {
		dvalue = defaultValue[0]
	}

	flag.StringVar(value, p.prefix+name, dvalue, "")
	p.keys = append(p.keys, name)
	p.vals = append(p.vals, value)
}

func (p *argsParser) Parse() {
	flag.Parse()

	fmt.Printf("Parsed args:\n")
	for i, key := range p.keys {
		val := p.vals[i]
		if v, ok := val.(*bool); ok && v != nil {
			val = *v
		} else if v, ok := val.(*int); ok && v != nil {
			val = *v
		} else if v, ok := val.(*int64); ok && v != nil {
			val = *v
		} else if v, ok := val.(*string); ok && v != nil {
			val = *v
		}

		fmt.Printf("    %s: %v\n", key, val)
	}
}
