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
			out <- nu.Value{Value: slices.Clone(k)}
			return nil
		})
	})
}

func listKeys(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}
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
				out <- nu.Value{Value: slices.Clone(k)}
			}
			return nil
		})
	})
}
