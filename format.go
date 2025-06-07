package main

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ainvaltin/nu-plugin"
)

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
