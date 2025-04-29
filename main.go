package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/ainvaltin/nu-plugin"
	"github.com/ainvaltin/nu-plugin/syntaxshape"
	"github.com/ainvaltin/nu-plugin/types"
)

//go:embed help.md
var longDesc string

func main() {
	p, err := nu.New(
		[]*nu.Command{boltCmd()},
		"0.0.1",
		nil,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create plugin", err)
		return
	}
	if err := p.Run(quitSignalContext()); err != nil && !errors.Is(err, nu.ErrGoodbye) {
		fmt.Fprintln(os.Stderr, "plugin exited with error", err)
	}
}

func boltCmd() *nu.Command {
	nameShape := syntaxshape.OneOf(syntaxshape.List(syntaxshape.Any()), syntaxshape.Binary(), syntaxshape.String())
	cmd := &nu.Command{
		Signature: nu.PluginSignature{
			Name:        "boltdb",
			Category:    "Database",
			Desc:        `Interact with bbolt database.`,
			Description: longDesc,
			SearchTerms: []string{"bbolt", "bolt"},
			InputOutputTypes: []nu.InOutTypes{
				{In: types.Nothing(), Out: types.Any()},
				{In: types.Binary(), Out: types.Any()},
				{In: types.String(), Out: types.Any()},
			},
			Named: []nu.Flag{
				{Long: "bucket", Short: "b", Shape: nameShape, Desc: "Name of the bucket to operate on. Nested buckets are represented by " +
					"list, ie path `foo -> bar` would be [foo, bar]. Nested lists can be used to build bucket name from parts. When not provided action takes place in the root bucket."},
				{Long: "key", Short: "k", Shape: nameShape, Desc: `Name of the key to operate on. If the value is List all items will be concatenated to single byte array, ie given '-k ["item " 0x[0005]]' the key name used would be string "item" followed by space and two bytes with values 0 and 5, it's equivalent to '-k 0x[6974656D200005]'.`},
				// potentially too confusing?
				//{Long: "path", Short: "p", Shape: nameShape, Desc: `Last item of the path is "key"`},
				{Long: "filter", Short: "f", Shape: syntaxshape.Closure(syntaxshape.Binary()), Desc: "Filter keys or buckets by name - closure is called for each key and if it returns `true` item is included in the output."},
			},
			RequiredPositional: nu.PositionalArgs{
				{Name: "file", Shape: syntaxshape.Filepath(), Desc: `Name of the Bolt database file.`},
				{Name: "action", Shape: syntaxshape.String(), Desc: "Operation to perform: buckets, keys, get, set, add, delete, stat, info"},
			},
			RestPositional:       &nu.PositionalArg{Name: "data", Shape: syntaxshape.OneOf(syntaxshape.Binary(), syntaxshape.String()), Desc: `Data for the operation, alternative for the input.`},
			AllowMissingExamples: true,
		},
		Examples: nu.Examples{
			{Description: `List root buckets`, Example: `boltdb /db/file.name buckets`, Result: &nu.Value{Value: []nu.Value{{Value: []byte{1, 2, 3, 4}}}}},
			{Description: `List buckets in the bucket "foo"`, Example: `boltdb /db/file.name buckets -b foo`, Result: &nu.Value{Value: []nu.Value{{Value: []byte("bar")}, {Value: []byte("zoo")}}}},
			{Description: `Save file content to a key "file.name" in the bucket "files" (read data from input)`, Example: `open /data/file.name --raw | boltdb /db/file.name set -b files -k file.name`},
			{Description: `Set key "buz" in nested bucket "foo -> bar" (read data from argument)`, Example: `boltdb /db/file.name set -b [foo, bar] -k buz 0x[010203]`},
			{Description: `List keys starting with bytes 0x[626c] (98 ja 108 in decimal or "bl" as string)`, Example: `boltdb /db/file.name keys -f {|| $in | bytes starts-with 0x[626c]}`, Result: &nu.Value{Value: []nu.Value{{Value: []byte{0x62, 0x6c, 111, 99, 107}}}}},
		},
		OnRun: boltCmdHandler,
	}
	return cmd
}

func boltCmdHandler(ctx context.Context, call *nu.ExecCommand) error {
	action, err := checkArgs(call)
	if err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	db, err := openDB(ctx, call, action)
	if err != nil {
		return fmt.Errorf("opening bolt db: %w", err)
	}
	defer db.Close()

	switch action {
	case "buckets":
		return listBuckets(ctx, db, call)
	case "keys":
		return listKeys(ctx, db, call)
	case "get":
		return getValue(ctx, db, call)
	case "set":
		return setValue(ctx, db, call)
	case "add":
		return addBucket(ctx, db, call)
	case "delete":
		return delete(ctx, db, call)
	case "stat":
		return stat(ctx, db, call)
	case "info":
		return info(ctx, db, call)
	default:
		return fmt.Errorf("unknown action %q", action)
	}
}

func checkArgs(call *nu.ExecCommand) (action string, err error) {
	_, path := call.FlagValue("path")
	_, bucket := call.FlagValue("bucket")
	_, key := call.FlagValue("key")
	if path && (bucket || key) {
		return "", fmt.Errorf(`when "path" flag is given "bucket" and/or "key" flag is not allowed`)
	}

	if len(call.Positional) == 3 && call.Input != nil {
		return "", fmt.Errorf(`both "data" argument and input can't be used at the same time`)
	}

	action = call.Positional[1].Value.(string)
	if action != "set" && (len(call.Positional) == 3 || call.Input != nil) {
		return "", fmt.Errorf(`action %q doesn't accept input`, action)
	}
	if (action == "get" || action == "set") && !key {
		return "", fmt.Errorf(`action %q requires "key" flag to be provided`, action)
	}
	if key && !slices.Contains([]string{"get", "set", "delete"}, action) {
		return "", fmt.Errorf(`action %q doesn't allow "key" flag`, action)
	}

	_, filter := call.FlagValue("filter")
	if filter && !slices.Contains([]string{"buckets", "keys"}, action) {
		return "", fmt.Errorf(`action %q doesn't support "filter" flag`, action)
	}

	return action, nil
}

func quitSignalContext() context.Context {
	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigChan)
		sig := <-sigChan
		cancel(fmt.Errorf("got quit signal: %s", sig))
	}()

	return ctx
}
