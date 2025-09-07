# Bolt database Nushell Plugin

[Nushell](https://www.nushell.sh/)
[Plugin](https://www.nushell.sh/contributor-book/plugins.html) 
to interact with [bbolt database](https://github.com/etcd-io/bbolt).

<img src="demo.gif" />

## Usage

Plugin implements `boltdb` command
```shell
boltdb {flags} <file> <action> ...(data)
```
ie to list all the buckets in the "blocks" bucket
```shell
boltdb /path/to/bbolt.db buckets -b blocks
```

See the [list of available actions](./help.md), to see the full help of the command run `boltdb --help`.

## Configuration

Configuration can be provided via `$env.config.plugins.NAME` Record with following keys:

| key | default | description |
|---|---|---|
| timeout | 3sec | Timeout for the open database call - only single process at a time may open bbolt database. |
| fileMode | 0600 | FileMode to use when opening database. |
| ReadOnly | false | If set to `true` databases are opened in read only mode, actions which modify the DB (`add`, `delete`, `set`) would then fail. |
| mustExist | false | If set to true database file must exist, otherwise plugin returns error. If both `ReadOnly` and `mustExist` are false `add` and `set` actions will create the database (if it doesn't exist, other actions still fail). |

See [bbolt documentation](https://pkg.go.dev/go.etcd.io/bbolt#Open) for more info about these parameters.

### Example configuration

Run `config env`, add

 ```
$env.config.plugins = {
    boltdb: {
        timeout: 5sec,
        ReadOnly: true
    }
}
```

save the file and reload the nushell configuration, ie

    source $nu.env-path

## Installation

Latest version is for [Nushell](https://www.nushell.sh/) version **0.107.0**.

Written in [Go](https://go.dev/) using 
[nu-plugin package](https://github.com/ainvaltin/nu-plugin).

To install it you have to have [Go installed](https://go.dev/dl/), then run
```sh
go install github.com/ainvaltin/nu_plugin_boltdb@latest
```
This creates the `nu_plugin_boltdb` binary in your `GOBIN` directory:

> Executables are installed in the directory named by the GOBIN environment
variable, which defaults to $GOPATH/bin or $HOME/go/bin if the GOPATH
environment variable is not set.

Locate the binary and follow instructions on 
[Downloading and installing a plugin](https://www.nushell.sh/book/plugins.html#downloading-and-installing-a-plugin)
page on how to register `nu_plugin_boltdb` as a plugin.
