package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"time"

	"go.etcd.io/bbolt"

	"github.com/ainvaltin/nu-plugin"
)

type configuration struct {
	timeout   time.Duration
	readOnly  bool
	fileMode  fs.FileMode
	mustExist bool // if true only existing files can be opened (ie can't create new DB)
}

func loadCfg(ctx context.Context, call *nu.ExecCommand) (configuration, error) {
	cfg := configuration{
		timeout:   3 * time.Second,
		readOnly:  false,
		fileMode:  0600,
		mustExist: false,
	}

	v, err := call.GetPluginConfig(ctx)
	if err != nil {
		return cfg, fmt.Errorf("reading plugin configuration: %w", err)
	}
	if v == nil {
		return cfg, nil
	}

	r, ok := v.Value.(nu.Record)
	if !ok {
		return cfg, fmt.Errorf("expected configuration to be Record, got %T", v.Value)
	}
	for k, v := range r {
		switch k {
		case "ReadOnly":
			cfg.readOnly = v.Value.(bool)
		case "timeout":
			cfg.timeout = v.Value.(time.Duration)
		case "fileMode":
			cfg.fileMode = fs.FileMode(v.Value.(int64))
		case "mustExist":
			cfg.mustExist = v.Value.(bool)
		}
	}
	return cfg, nil
}

func openDB(ctx context.Context, call *nu.ExecCommand, action string) (*bbolt.DB, error) {
	cfg, err := loadCfg(ctx, call)
	if err != nil {
		return nil, err
	}

	dbName := call.Positional[0].Value.(string)
	if _, err := os.Stat(dbName); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if cfg.mustExist || !slices.Contains([]string{"add", "set"}, action) {
				return nil, fmt.Errorf("database does not exist and creating databases is disabled")
			}
		} else {
			return nil, fmt.Errorf("invalid database name: %w", err)
		}
	}

	db, err := bbolt.Open(dbName, cfg.fileMode, &bbolt.Options{Timeout: cfg.timeout, ReadOnly: cfg.readOnly})
	if err != nil {
		return nil, fmt.Errorf("opening bolt db: %w", err)
	}
	return db, nil
}
