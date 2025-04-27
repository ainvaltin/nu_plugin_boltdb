# Bolt database Nushell Plugin

[Nushell](https://www.nushell.sh/)
[Plugin](https://www.nushell.sh/contributor-book/plugins.html) 
to interact with [bbolt database](https://github.com/etcd-io/bbolt).

Written in [Go](https://go.dev/) using 
[nu-plugin package](https://github.com/ainvaltin/nu-plugin).

## Usage

Plugin implements `boltdb` command
```shell
boltdb {flags} <file> <action> ...(data)
```
ie to list all the buckets in the "blocks" bucket
```shell
boltdb /path/to/bbolt.db buckets -b blocks
```

To see the full help of the command run `boltdb --help`.


## Installation

Latest version is for Nushell version `0.103.0`.

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
