Lesson: Install ipfs-pack
=====

## Prerequisites

* Install [go](https://golang.org/dl/)

You do not need to install ipfs separately. ipfs-pack will handle that for you.

## (SOON) Download the Prebuilt Binaries

_If you see ipfs-pack binaries listed on dist.ipfs.io, then this document is out of date. Please submit a PR or an issue on Github to correct it._

Currently there are not prebuilt ipfs-pack binaries. After we have tested the tool more thoroughly we will release prebuilt binaries on dist.ipfs.io.  In the meantime, you will have to build from source. (read on...)

## Build from source

### Step 1: Clone the Repository

```sh
git clone git@github.com:ipfs/ipfs-pack
cd ipfs-pack
```

### Step 2: Build the binaries from source
Build ipfs-pack, which includes go-ipfs. This will take a while because it downloads and builds all of go-ipfs.
```
make build
```

This generates a binary called `ipfs-pack`

### Step 3: add the binary to your PATH

Add the generated binary to your executable PATH. The way to do this depends on your operating system.


## Next Steps

Next, [Initialize a Pack](initialize-a-pack.md)
