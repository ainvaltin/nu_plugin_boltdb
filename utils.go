package main

import (
	"fmt"
	"slices"
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
				return nil, fmt.Errorf("invalid path, bucket %s does not contain bucket %s", pathStr(path[:i]), pathStr([][]byte{v}))
			}
			return nil, fmt.Errorf("invalid path, root bucket does not contain bucket %s", pathStr([][]byte{v}))
		}
	}
	return root, nil
}

func pathStr(path [][]byte) string {
	s := ""
	for _, v := range path {
		if n := formatName(v, false); len(n) == 1 {
			s += n[0] + " -> "
		} else {
			s += "[" + strings.Join(n, ", ") + "] -> "
		}
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

func stringifyName(name []byte) nu.Value {
	r := formatName(name, true)
	if len(r) == 1 {
		return nu.Value{Value: r[0]}
	}

	return nu.Value{Value: "[" + strings.Join(r, ", ") + "]"}
}

func textName(name []byte) nu.Value {
	r := formatName(name, false)
	if len(r) == 1 {
		return nu.Value{Value: r[0]}
	}

	return nu.Value{Value: "[" + strings.Join(r, ", ") + "]"}
}

func formatName(name []byte, stringify bool) []string {
	r := tokenizeName(slices.Clone(name), stringify)
	s := make([]string, 0, len(r))
	for _, v := range r {
		if t, ok := v.(string); ok {
			// do we need to quote it?
			s = append(s, t)
		} else {
			s = append(s, fmt.Sprintf("0x[%x]", v))
		}
	}
	return s
}

func tokenizeName(name []byte, stringify bool) []any {
	r := make([]any, 0, 1)
	printable := false // is the last run in "r" printable
	for i := 0; i < len(name); {
		size, flags := printableRun(name[i:])
		if size > 0 {
			// minimum str len configuration?
			if size < 3 && len(r) > 0 {
				lr := r[len(r)-1].([]byte)
				r[len(r)-1] = append(lr, name[i:i+size]...)
			} else {
				s := string(name[i : i+size])
				if stringify {
					switch {
					case flags == 0: // OK to use bare string
					case flags&flagSQuote == 0:
						s = "'" + s + "'"
					case flags&flagBacktick == 0:
						s = "`" + s + "`"
					case flags&flagDQuote == 0 && flags&flagBackslash == 0:
						s = `"` + s + `"`
					default:
						s = fmt.Sprintf("%q", s)
					}
				}
				r = append(r, s)
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
	return r
}

func printableRun(b []byte) (idx int, flags uint32) {
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError || !unicode.IsPrint(r) {
			return i, flags
		}
		switch r {
		case '"':
			flags |= flagDQuote
		case '\'':
			flags |= flagSQuote
		case '`':
			flags |= flagBacktick
		case '\\':
			flags |= flagBackslash
		case ' ':
			flags |= flagSpace
		}

		i += size
	}
	return len(b), flags
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

const (
	flagDQuote = 1 << iota
	flagSQuote
	flagBacktick
	flagBackslash
	flagSpace
)
