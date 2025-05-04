package main

import (
	"context"
	"fmt"
	"regexp"
	"slices"

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

	format := func(name []byte) nu.Value { return nu.Value{Value: name} }
	if v, _ := call.FlagValue("stringify"); v.Value.(bool) {
		format = formatName
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
			if filter(k) {
				out <- format(slices.Clone(k))
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

	format := func(name []byte) nu.Value { return nu.Value{Value: name} }
	if v, _ := call.FlagValue("stringify"); v.Value.(bool) {
		format = formatName
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
			if v != nil && filter(k) {
				out <- format(slices.Clone(k))
			}
			return nil
		})
	})
}

func getFilter(call *nu.ExecCommand) (func(key []byte) bool, error) {
	match, ok := call.FlagValue("match")
	if !ok {
		return func([]byte) bool { return true }, nil
	}

	reg, err := regexp.Compile(match.Value.(string))
	if err != nil {
		return nil, fmt.Errorf("compiling regular expression: %w", err)
	}
	return func(key []byte) bool { return reg.Match(key) }, nil
}
