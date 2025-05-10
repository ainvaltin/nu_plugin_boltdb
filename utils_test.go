package main

import (
	"testing"
)

func Test_formatName(t *testing.T) {
	t.Run("all printable", func(t *testing.T) {
		var testCases = []struct {
			in  []byte
			out string
		}{
			{in: []byte(`str`), out: `str`},
			{in: []byte(`foo bar`), out: `'foo bar'`},
			{in: []byte(`quote"`), out: `'quote"'`},
			{in: []byte(`A B`), out: `'A B'`},
			// quotes inside string - should be escaped?
			{in: []byte(`A"B`), out: `'A"B'`},
			{in: []byte(`A'B`), out: "`A'B`"},
			{in: []byte(`A'B"`), out: "`A'B\"`"},
			{in: []byte(" ' ` \" "), out: "\" ' ` \\\" \""},
			{in: []byte(" ' ` \" \\ "), out: "\" ' ` \\\" \\\\ \""},
			{in: []byte{' ', '\'', ' ', '`', ' ', '"', ' ', '\\', ' '}, out: "\" ' ` \\\" \\\\ \""},
		}

		for i, tc := range testCases {
			v := formatName(tc.in)
			s, ok := v.Value.(string)
			if !ok {
				t.Errorf("[%d] expected Value to be string, got %T", i, v.Value)
			}
			if s != tc.out {
				t.Errorf("[%d] expected %q, got %q", i, tc.out, s)
			}
		}
	})

	t.Run("all binary", func(t *testing.T) {
		var testCases = []struct {
			in  []byte
			out string
		}{
			{in: nil, out: `[]`}, // name can't be empty so we do not worry about this corner case
			{in: []byte{}, out: `[]`},
			{in: []byte{0}, out: `0x[00]`},
			{in: []byte{0, 1}, out: `0x[0001]`},
			{in: []byte{0, 1, 2}, out: `0x[000102]`},
			{in: []byte{128}, out: `0x[80]`},
			{in: []byte{128, 127, 126, 125}, out: `0x[807f7e7d]`},
		}

		for i, tc := range testCases {
			v := formatName(tc.in)
			s, ok := v.Value.(string)
			if !ok {
				t.Errorf("[%d] expected Value to be string, got %T", i, v.Value)
			}
			if s != tc.out {
				t.Errorf("[%d] expected %q, got %q", i, tc.out, s)
			}
		}
	})

	t.Run("mixed binary and printable", func(t *testing.T) {
		var testCases = []struct {
			in  []byte
			out string
		}{
			// starting with string
			{in: append([]byte(`str`), 0, 0), out: `[str, 0x[0000]]`},
			{in: append([]byte(`A`), 0, 0), out: `[A, 0x[0000]]`},
			{in: append([]byte(`B`), 255, 254, 'A'), out: `[B, 0x[fffe41]]`},
			{in: append([]byte(`C`), 255, 254, 'A', 'c'), out: `[C, 0x[fffe4163]]`},
			{in: append([]byte(`D`), 255, 254, 'A', 'B', 'd'), out: `[D, 0x[fffe], ABd]`},
			// starting with binary
			{in: append([]byte{0, 1}, 'A'), out: `0x[000141]`},
			{in: append([]byte{0, 1}, 'A', 'B'), out: `0x[00014142]`},
			{in: append([]byte{0, 1}, 'A', 'B', 'C'), out: `[0x[0001], ABC]`},
			// string between binary
			{in: append([]byte{0, 1}, 'A', 3, 4), out: `0x[0001410304]`},
			{in: append([]byte{1, 2}, 'A', 'B', 3, 4), out: `0x[010241420304]`},
			{in: append([]byte{2, 3}, 'A', 'B', 'C', 4, 5), out: `[0x[0203], ABC, 0x[0405]]`},
			// binary between strings
			{in: append([]byte(`A`), 0, 0x11, 'A'), out: `[A, 0x[001141]]`},
			{in: append([]byte(`A`), 0, 0x11, 'A', 'B'), out: `[A, 0x[00114142]]`},
			{in: append([]byte(`A`), 0, 0x11, 'A', 'B', 'C'), out: `[A, 0x[0011], ABC]`},
		}

		for i, tc := range testCases {
			v := formatName(tc.in)
			s, ok := v.Value.(string)
			if !ok {
				t.Errorf("[%d] expected Value to be string, got %T", i, v.Value)
			}
			if s != tc.out {
				t.Errorf("[%d] expected %q, got %q", i, tc.out, s)
			}
		}
	})
}
