package main

import (
	"errors"
	"fmt"

	"github.com/ainvaltin/nu-plugin"
)

func toBytes(v nu.Value) ([]byte, error) {
	switch t := v.Value.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	case int64:
		if t < 256 {
			return []byte{uint8(t)}, nil
		}
		return nil, &nu.Error{
			Err:    fmt.Errorf("integer values must fit into byte, got %d", t),
			Labels: []nu.Label{{Text: "value out of range (max allowed 255)", Span: v.Span}},
		}
	case []nu.Value:
		var r []byte
		for _, v := range t {
			b, err := toBytes(v)
			if err != nil {
				return nil, err
			}
			r = append(r, b...)
		}
		return r, nil
		// IntRange - generates binary blob?
		// Closure - returns binary?
	default:
		return nil, nu.Error{
			Err:    errors.New("can't convert value to bytes"),
			Help:   "Supported types are Binary and String",
			Labels: []nu.Label{{Text: fmt.Sprintf("unsupported type %T", t), Span: v.Span}},
		}
	}
}
