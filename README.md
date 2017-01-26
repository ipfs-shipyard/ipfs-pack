# ipfs-pack - filesystem packing tool

`ipfs-pack` is a tool and library to work with ipfs and large collections of data in UNIX/POSIX filesystems.

- It identifies singular collections or bundles of data (the pack).
- It creates a light-weight cryptographically secure manifest that preserves the integrity of the collection over time, and _travels with the data_ (PackManifest).
- It helps use ipfs in a mode that references the filesystem files directly and avoids duplicating data (filestore).
- It carries a standard dataset metadata file to capture and present information about the dataset (data-package.json).
- It helps verify the authenticity of data through a file carrying cryptographic signatures (PackAuth).


## Installing

With go installed, simply run `make build`.

## Spec

Read the `ipfs-pack` work-in-progress "spec" here: [Spec (WIP)](./spec.md).

