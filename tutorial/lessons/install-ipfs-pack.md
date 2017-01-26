Lesson: Install ipfs-pack
=====

## Prerequisites

* Install [go]

You do not need to install ipfs separately. ipfs-pack will handle that for you.

## Build from source

```sh
git clone git@github.com:ipfs/ipfs-pack
cd ipfs-pack
```

Build ipfs-pack, which includes go-ipfs. This will take a while because it downloads and builds all of go-ipfs.
```
make build
```
