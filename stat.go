package main

import (
	"context"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func stat(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}
	if len(path) == 0 {
		return call.ReturnValue(ctx, nu.ToValue(db.Stats()))
	}

	return db.View(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx, path)
		if err != nil {
			return err
		}
		return call.ReturnValue(ctx, nu.ToValue(b.Stats()))
	})
}

func info(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}
	return db.View(func(tx *bbolt.Tx) error {
		b, err := goToBucket(tx, path)
		if err != nil {
			return err
		}
		return call.ReturnValue(ctx, nu.ToValue(b.Inspect()))
	})
}
