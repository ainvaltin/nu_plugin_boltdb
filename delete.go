package main

import (
	"context"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

func delete(ctx context.Context, db *bbolt.DB, call *nu.ExecCommand) error {
	path, key, err := location(call)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		if key == nil {
			b, err := goToBucket(tx.Cursor().Bucket(), path[:len(path)-1])
			if err != nil {
				return err
			}
			return b.DeleteBucket(path[len(path)-1])
		}

		b, err := goToBucket(tx.Cursor().Bucket(), path)
		if err != nil {
			return err
		}
		return b.Delete(key)
	})
}
