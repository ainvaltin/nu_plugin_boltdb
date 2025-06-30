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

func (cfg *configuration) parse(v nu.Value) error {
	r, ok := v.Value.(nu.Record)
	if !ok {
		return nu.Error{Err: fmt.Errorf("expected configuration to be Record, got %T", v.Value), Labels: []nu.Label{{Text: "expected Record", Span: v.Span}}}
	}
	for k, v := range r {
		switch k {
		case "ReadOnly":
			if cfg.readOnly, ok = v.Value.(bool); !ok {
				return expectedBool("ReadOnly", v)
			}
		case "timeout":
			if cfg.timeout, ok = v.Value.(time.Duration); !ok {
				return nu.Error{
					Err:    fmt.Errorf("expected 'timeout' to be Duration, got %T", v.Value),
					Help:   "Duration is a number followed by unit, ie 5sec",
					Labels: []nu.Label{{Text: "expected Duration", Span: v.Span}},
				}
			}
		case "fileMode":
			var m int64
			if m, ok = v.Value.(int64); !ok {
				return nu.Error{
					Err:    fmt.Errorf("expected 'fileMode' to be integer, got %T", v.Value),
					Labels: []nu.Label{{Text: "expected Integer", Span: v.Span}},
				}
			}
			cfg.fileMode = fs.FileMode(m)
		case "mustExist":
			if cfg.mustExist, ok = v.Value.(bool); !ok {
				return expectedBool("mustExist", v)
			}
		}
	}
	return nil
}

func expectedBool(name string, v nu.Value) error {
	return nu.Error{
		Err:    fmt.Errorf("expected %q to be boolean, got %T", name, v.Value),
		Help:   "Valid values are 'true' and 'false' (without quotes).",
		Labels: []nu.Label{{Text: "expected boolean", Span: v.Span}},
	}
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

	if err = cfg.parse(*v); err != nil {
		return cfg, fmt.Errorf("parsing plugin configuration: %w", err)
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
				return nil, nu.Error{
					Err:    fmt.Errorf("database does not exist"),
					Code:   "boltdb::config::mustExist",
					Url:    "https://github.com/ainvaltin/nu_plugin_boltdb?tab=readme-ov-file#configuration",
					Help:   `Only "add" and "set" actions are allowed to create database as the "mustExist" configuration flag is set to "true".`,
					Labels: []nu.Label{{Text: "file does not exist", Span: call.Positional[0].Span}},
				}
			}
		} else {
			return nil, nu.Error{Err: fmt.Errorf("invalid database name: %w", err), Labels: []nu.Label{{Text: err.Error(), Span: call.Positional[0].Span}}}
		}
	}

	db, err := bbolt.Open(dbName, cfg.fileMode, &bbolt.Options{Timeout: cfg.timeout, ReadOnly: cfg.readOnly})
	if err != nil {
		return nil, fmt.Errorf("opening bolt db: %w", err)
	}
	return db, nil
}
