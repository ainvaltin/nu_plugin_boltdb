package main

import (
	"context"
	"fmt"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func listBuckets(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}

	filter, err := getFilter(call)
	if err != nil {
		return err
	}

	format := getFormatter(call)

	return db.View(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx, path)
		if err != nil {
			return err
		}

		out, err := call.ReturnListStream(ctx)
		if err != nil {
			return fmt.Errorf("creating result stream: %w", err)
		}
		defer close(out)

		return b.ForEachBucket(func(k []byte) error {
			if filter(k) {
				out <- format(k)
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

	filter, err := getFilter(call)
	if err != nil {
		return err
	}

	format := getFormatter(call)

	return db.View(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx, path)
		if err != nil {
			return err
		}

		out, err := call.ReturnListStream(ctx)
		if err != nil {
			return fmt.Errorf("creating result stream: %w", err)
		}
		defer close(out)

		return b.ForEach(func(k, v []byte) error {
			if v != nil && filter(k) {
				out <- format(k)
			}
			return nil
		})
	})
}
