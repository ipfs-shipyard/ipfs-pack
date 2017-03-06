# ipfs-pack - filesystem packing tool

`ipfs-pack` is a tool and library to work with ipfs and large collections of data in UNIX/POSIX filesystems.

- It identifies singular collections or bundles of data (the pack).
- It creates a light-weight cryptographically secure manifest that preserves the integrity of the collection over time, and _travels with the data_ (PackManifest).
- It helps use ipfs in a mode that references the filesystem files directly and avoids duplicating data (filestore).
- It carries a standard dataset metadata file to capture and present information about the dataset (data-package.json).
- It helps verify the authenticity of data through a file carrying cryptographic signatures (PackAuth).


## Installing

Pre-built binaries are available on the [ipfs distributions page](https://dist.ipfs.io/#ipfs-pack).

### From source
If there is not a pre-built binary for your system, or you'd like to try out
unreleased features, or for any other reason you want to build from source, its
relatively simple.  First, make sure you have go installed and properly
configured. [This guide](https://golang.org/doc/install) from the go team
should help with that.  Once thats done, simply run `make build`.

## Usage

```
$ ipfs-pack --help
NAME:
   ipfs-pack - A filesystem packing tool.

USAGE:
   ipfs-pack [global options] command [command options] [arguments...]

VERSION:
   v0.1.0

COMMANDS:
     make     makes the package, overwriting the PackManifest file.
     verify   verifies the ipfs-pack manifest file is correct.
     repo     manipulate the ipfs repo cache associated with this pack.
     serve    start an ipfs node to serve this pack's contents.
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Make a pack
```
$ cd /path/to/data/dir
$ ipfs-pack make
wrote PackManifest
```

### Verify a pack
```
$ ipfs-pack verify
Pack verified successfully!
```

## Testing
Tests require the [random-files](https://github.com/jbenet/go-random-files) module

```bash
go get -u github.com/jbenet/go-random-files/random-files
```

Run tests with
```bash
./test/pack-basic.sh
./test/pack-serve.sh
```

## Spec

Read the `ipfs-pack` work-in-progress "spec" here: [Spec (WIP)](./spec.md).

