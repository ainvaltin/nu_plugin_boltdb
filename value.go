package main

import (
	"context"
	"fmt"
	"io"
	"slices"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func addBucket(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(path[0])
		if err != nil {
			return err
		}
		for _, v := range path[1:] {
			if b, err = b.CreateBucketIfNotExists(v); err != nil {
				return err
			}
		}
		return nil
	})
}

func getValue(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, key, err := location(call)
	if err != nil {
		return err
	}
	filter, err := getFilter(call)
	if err != nil {
		return err
	}

	return db.View(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx.Cursor().Bucket(), path)
		if err != nil {
			return err
		}

		if key != nil {
			if v := b.Get(key); v != nil {
				return call.ReturnValue(ctx, nu.Value{Value: slices.Clone(v)})
			}
			return nil
		}

		out, err := call.ReturnListStream(ctx)
		if err != nil {
			return fmt.Errorf("creating result stream: %w", err)
		}
		defer close(out)

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v != nil && filter(k) {
				out <- nu.Value{Value: nu.Record{
					"key":   nu.Value{Value: slices.Clone(k)},
					"value": nu.Value{Value: slices.Clone(v)},
				}}
			}
		}
		return nil
	})
}

func setValue(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, key, err := location(call)
	if err != nil {
		return err
	}
	v, err := inputValue(call)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx.Cursor().Bucket(), path)
		if err != nil {
			return err
		}
		return b.Put(key, v)
	})
}

func inputValue(call *nu.ExecCommand) ([]byte, error) {
	if len(call.Positional) == 3 {
		return toBytes(call.Positional[2])
	}

	switch in := call.Input.(type) {
	case nil:
		return nil, fmt.Errorf("input value is missing")
	case nu.Value:
		return toBytes(in)
	case io.ReadCloser:
		return io.ReadAll(in)
	default:
		return nil, fmt.Errorf("unsupported input type %T", call.Input)
	}
}
