package main

import (
	"fmt"
	"strings"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func goToBucket(root *bbolt.Bucket, path [][]byte) (*bbolt.Bucket, error) {
	for i, v := range path {
		if root = root.Bucket(v); root == nil {
			return nil, fmt.Errorf("invalid path, bucket %q does not contain bucket %x", pathStr(path[:i]), v)
		}
	}
	return root, nil
}

func pathStr(path [][]byte) string {
	s := ""
	for _, v := range path {
		s += fmt.Sprintf("%x -> ", v)
	}
	return strings.TrimSuffix(s, " -> ")
}

// figure out from flags the path and key of the request
func location(call *nu.ExecCommand) (bucket [][]byte, key []byte, err error) {
	// do we have "path" which combines bucket and key?
	if b, ok := call.FlagValue("path"); ok {
		bucket, err = toPath(b)
		return bucket[:len(bucket)-1], bucket[len(bucket)-1], err
	}

	// do we have separate bucket and key params?
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
