package main

import (
	"context"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func addBucket(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, _, err := location(call)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(path[0].name)
		if err != nil {
			return err
		}
		for _, v := range path[1:] {
			if b, err = b.CreateBucketIfNotExists(v.name); err != nil {
				return err
			}
		}
		return nil
	})
}
