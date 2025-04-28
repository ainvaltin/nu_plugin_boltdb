package main

import (
	"context"
	"fmt"
	"slices"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func listBuckets(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}
	filter := getFilter(call)

	return db.View(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx.Cursor().Bucket(), path)
		if err != nil {
			return err
		}
		out, err := call.ReturnListStream(ctx)
		if err != nil {
			return fmt.Errorf("creating result stream: %w", err)
		}
		defer close(out)

		return b.ForEachBucket(func(k []byte) error {
			if ok, err := filter(ctx, k); ok {
				out <- nu.Value{Value: slices.Clone(k)}
			} else if err != nil {
				return fmt.Errorf("evaluating filter closure: %w", err)
			}
			return nil
		})
	})
}

func listKeys(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}
	filter := getFilter(call)

	return db.View(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx.Cursor().Bucket(), path)
		if err != nil {
			return err
		}
		out, err := call.ReturnListStream(ctx)
		if err != nil {
			return fmt.Errorf("creating result stream: %w", err)
		}
		defer close(out)

		return b.ForEach(func(k, v []byte) error {
			if v != nil {
				if ok, err := filter(ctx, k); ok {
					out <- nu.Value{Value: slices.Clone(k)}
				} else if err != nil {
					return fmt.Errorf("evaluating filter closure: %w", err)
				}
			}
			return nil
		})
	})
}

func getFilter(call *nu.ExecCommand) func(ctx context.Context, key []byte) (bool, error) {
	closure, ok := call.FlagValue("filter")
	if !ok {
		return func(context.Context, []byte) (bool, error) { return true, nil }
	}

	return func(ctx context.Context, key []byte) (bool, error) {
		r, err := call.EvalClosure(ctx, closure, nu.InputValue(nu.Value{Value: key}))
		if err != nil {
			return false, fmt.Errorf("evaluating filter closure: %w", err)
		}
		b, ok := r.(nu.Value)
		if !ok {
			return false, fmt.Errorf("expected that filter closure returns single Value, got %T = %v", r, r)
		}
		v, ok := b.Value.(bool)
		if !ok {
			return false, fmt.Errorf("expected that filter closure returns bool, got %T = %v", r, r)
		}
		return v, nil
	}
}
