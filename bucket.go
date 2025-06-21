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
		b := tx.Cursor().Bucket()
		for _, v := range path {
			if b, err = b.CreateBucketIfNotExists(v.name); err != nil {
				return nu.Error{
					Err:    err,
					Labels: []nu.Label{{Text: "invalid bucket", Span: v.span}},
				}
			}
		}
		return nil
	})
}
