package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func goToBucket(root *bbolt.Bucket, path [][]byte) (*bbolt.Bucket, error) {
	for i, v := range path {
		if root = root.Bucket(v); root == nil {
			if i > 0 {
				return nil, fmt.Errorf("invalid path, bucket %q does not contain bucket 0x[%x]", pathStr(path[:i]), v)
			}
			return nil, fmt.Errorf("invalid path, root bucket does not contain bucket 0x[%x]", v)
		}
	}
	return root, nil
}

func pathStr(path [][]byte) string {
	s := ""
	for _, v := range path {
		s += fmt.Sprintf("0x[%x] -> ", v)
	}
	return strings.TrimSuffix(s, " -> ")
}

// figure out from flags the path and key of the request
func location(call *nu.ExecCommand) (bucket [][]byte, key []byte, err error) {
	if b, ok := call.FlagValue("bucket"); ok {
		if bucket, err = toPath(b); err != nil {
			return nil, nil, err
		}
	}
	if b, ok := call.FlagValue("key"); ok {
		key, err = toBytes(b)
	}
	return bucket, key, err
}

func toPath(v nu.Value) ([][]byte, error) {
	switch t := v.Value.(type) {
	case []nu.Value:
		var r [][]byte
		for _, v := range t {
			b, err := toBytes(v)
			if err != nil {
				return nil, err
			}
			r = append(r, b)
		}
		return r, nil
	default:
		b, err := toBytes(v)
		return [][]byte{b}, err
	}
}

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
		return nil, fmt.Errorf("integer values must fit into byte, got %d", t)
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
	default:
		return nil, fmt.Errorf("unsupported type %T", t)
	}
}

func formatName(name []byte) nu.Value {
	r := make([]any, 0, 1)
	printable := false // is the last run in "r" printable
	for i := 0; i < len(name); {
		size := printableRun(name[i:])
		if size > 0 {
			// minimum str len configuration?
			if size < 3 && len(r) > 0 {
				lr := r[len(r)-1].([]byte)
				r[len(r)-1] = append(lr, name[i:i+size]...)
			} else {
				r = append(r, string(name[i:i+size]))
				printable = true
			}
		}

		if i += size; i == len(name) {
			break
		}

		// followed by unprintable
		size = unprintableRun(name[i:])
		if printable || len(r) == 0 {
			// if the printable item is shorter than minimum convert to binary?
			r = append(r, name[i:i+size])
			printable = false
		} else {
			lr := r[len(r)-1].([]byte)
			r[len(r)-1] = append(lr, name[i:i+size]...)
		}
		i += size
	}

	if len(r) == 1 {
		if printable {
			return nu.Value{Value: r[0]}
		}
		return nu.Value{Value: fmt.Sprintf("0x[%x]", r[0])}
	}

	s := "["
	for _, v := range r {
		switch t := v.(type) {
		case string:
			s += fmt.Sprintf("%q, ", t)
		case []byte:
			s += fmt.Sprintf("0x[%x], ", t)
		}
	}
	return nu.Value{Value: strings.TrimSuffix(s, ", ") + "]"}
}

func printableRun(b []byte) int {
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError || !unicode.IsPrint(r) {
			return i
		}
		i += size
	}
	return len(b)
}

func unprintableRun(b []byte) int {
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if !(r == utf8.RuneError || !unicode.IsPrint(r)) {
			return i
		}
		i += size
	}
	return len(b)
}
