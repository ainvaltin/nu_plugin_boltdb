package main

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/ainvaltin/nu-plugin"
)

func location(call *nu.ExecCommand) (bucket []boltItem, key *boltItem, err error) {
	if b, ok := call.FlagValue("bucket"); ok {
		if bucket, err = toPath(b); err != nil {
			return nil, nil, fmt.Errorf("invalid bucket name: %w", err)
		}
	}

	if b, ok := call.FlagValue("key"); ok {
		k, err := toBytes(b)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid key name: %w", err)
		}
		key = &boltItem{name: k, span: b.Span}
	}
	return bucket, key, nil
}

func getFilter(call *nu.ExecCommand) (func(key []byte) bool, error) {
	match, ok := call.FlagValue("match")
	if !ok {
		return func([]byte) bool { return true }, nil
	}

	reg, err := regexp.Compile(match.Value.(string))
	if err != nil {
		return nil, nu.Error{
			Err:    fmt.Errorf("compiling regular expression: %w", err),
			Url:    "https://pkg.go.dev/regexp/syntax",
			Help:   "See Go documentation about supported regular expression syntax",
			Labels: []nu.Label{{Text: "invalid regexp", Span: match.Span}},
		}
	}
	return func(key []byte) bool { return reg.Match(key) }, nil
}

func getFormatter(call *nu.ExecCommand) func([]byte) nu.Value {
	// the default is native/binary format
	format := func(name []byte) nu.Value { return nu.Value{Value: slices.Clone(name)} }

	fmtFlag, ok := call.FlagValue("format")
	if !ok {
		return format
	}
	switch fmtFlag.Value.(string) {
	case "stringify":
		return stringifyName
	case "text":
		return textName
	case "hex":
		return func(b []byte) nu.Value { return nu.Value{Value: fmt.Sprintf("%x", b)} }
	case "HEX":
		return func(b []byte) nu.Value { return nu.Value{Value: fmt.Sprintf("%X", b)} }
	}
	return format
}
