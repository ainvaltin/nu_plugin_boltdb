package main

import (
	"errors"
	"fmt"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

/*
boltItem is bucket or key name in bbolt database.
*/
type boltItem struct {
	name []byte
	span nu.Span
}

func toPath(v nu.Value) (path []boltItem, _ error) {
	switch t := v.Value.(type) {
	case []nu.Value:
		for _, v := range t {
			b, err := toBytes(v)
			if err != nil {
				return nil, err
			}
			path = append(path, boltItem{name: b, span: v.Span})
		}
		return path, nil
	case nu.CellPath:
		for _, v := range t.Members {
			if v.Type() != nu.PathVariantString {
				return nil, (&nu.Error{Err: errors.New("only string path members are supported")}).AddLabel("integer members not supported", v.Span())
			}
			if v.Optional() {
				// support optional items as last member(s)?
				return nil, (&nu.Error{Err: errors.New("optional path members are not supported")}).AddLabel("optional members not supported", v.Span())
			}
			path = append(path, boltItem{name: []byte(v.PathStr()), span: v.Span()})
		}
		return path, nil
	default:
		b, err := toBytes(v)
		return []boltItem{{name: b, span: v.Span}}, err
	}
}

func goToBucket(tx *bbolt.Tx, path []boltItem) (*bbolt.Bucket, error) {
	b := tx.Cursor().Bucket()
	for _, v := range path {
		if b = b.Bucket(v.name); b == nil {
			return nil, (&nu.Error{Err: fmt.Errorf("bucket %x doesn't exist", v.name)}).AddLabel("no such bucket", v.span)
		}
	}
	return b, nil
}
