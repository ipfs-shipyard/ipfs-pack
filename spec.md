# ipfs-pack spec

## Background

Please read the original proposal here: https://github.com/ipfs/notes/issues/205, and the notes describing user stories in this repository: https://github.com/ipfs/archives/issues?utf8=%E2%9C%93&q=label%3Aipfs-pack%20is%3Aissue

`ipfs-pack` is meant to be a bridge between POSIX and the world of content-addressed hash-linked data (IPFS and IPLD), that increases the safety and integrity of data in POSIX filesystems while avoiding duplicating or moving data into an ipfs repository. `ipfs-pack` establishes the notion of a "pack" of files (like a bundle or bag. We use `pack` to avoid confusing it with a `Bag` from `BagIt`, a similar format that `ipfs-pack` is compatible with).

### Status

`ipfs-pack` is only a spec at present, but the hope is to implement it imminently.

## Abstractions

### Terms

- **pack**: a bundle of directories, files, and metadata that represents a single collection of data. A pack is cryptographically secured with hashing (forms a merkle-dag). It is useful to ease working with large collections of data that want to be: secured from tampering and bitrot, certified by authors, and distributed across machines.
- **pack tree**: the filesystem subtree (directories and files) that represents an ipfs-pack.
- **pack root**: the root directory of the **pack tree**. it contains everything.
- **pack contents**: the directories and files that are "contained" in the pack (no pack metadata).
- **PackManifest**: a file that contains a listing of all entries in the pack tree, along with their cryptographic hash. It is a lightweight way to ensure integrity of the whole pack, and to verify what items may have been modified.
- **pack repo**: an ipfs object repository for a pack. It contains and can retrieve all ipld objects within the pack. It can be considered a cache: as long as the **pack tree** itself remains intact, the **pack repo** can be destroyed safely and reconstructed at any point. The **pack repo** helps run ipfs commands on the pack, for example serve and distribute the pack to the rest of the world. A **pack repo** SHOULD be using **filestore**, which means the raw data of a pack will be referenced and not copied.
- **filestore**: an ipfs repo storage strategy that only keeps references to raw file data, instead of copying all the data into another storage engine within the ipfs repo. This is useful when working with massive archives to avoid duplicating the data when importing it into ipfs.

### Diagram

```
+---- unixy filesystem -----------------------------------------------------+
|                                                                           |
|  /a/b/c/d/pack-root-dir                                                   |
|                                                                           |
|   +---- ipfs pack +---------------------------------------------------+   |
|   |                                                                   |   |
|   |  +-------------------------------------------------------------+  |   |
|   |  |                                                             |  |   |
|   |  |                                                             |  |   |
|   |  |       contents - files and dirs contained in the pack       |  |   |
|   |  |  +---->                                             <----+  |  |   |
|   |  |  |                                                       |  |  |   |
|   |  +-------------------------------------------------------------+  |   |
|   |     |                                                       |     |   |
|   |     |                                                       |     |   |
|   |     |                                                       |     |   |
|   |  +--+-------------+  +--------------------+  +--------------+--+  |   |
|   |  |                |  |                    |  |                 |  |   |
|   |  |  PackManifest  |  |      PackAuth      |  |   .ipfs-pack    |  |   |
|   |  | (secure index) |  |  (authentication)  |  |  (object repo)  |  |   |
|   |  |     (MUST)     |  |       (MAY)        |  |      (MAY)      |  |   |
|   |  +------------+---+  +---+----------------+  +-----------------+  |   |
|   |               ^          |                                        |   |
|   |               |          |                                        |   |
|   |               +----------+                                        |   |
|   +-------------------------------------------------------------------+   |
|                                                                           |
+---------------------------------------------------------------------------+
```

### Listing

This is an example listing of a pack

```
> tree /a/b/c/d/pack-root-dir
├── .ipfs         <--- ipfs object repo (optional, cache)
├── PackAuth      <--- cryptographic signatures (optional)
├── PackManifest  <--- cryptographic hash manifest (required)
└── foo           <--- files and dirs contained in the pack
    ├── bar
    │   ├── 1
    │   └── 2
    └── baz
        ├── 2
        └── 3
```


## Commands

```
> ipfs-pack -h
USAGE
    ipfs-pack <subcommand> <arguments>

SUBCOMMANDS
    make     makes the package, overwriting the PackManifest file.
    verify   verifies the ipfs-pack manifest file is correct.
    status   shows the status of changes in the pack
    repo     creates (or updates) a temporary ipfs object repo at `.ipfs-pack`
    serve    starts an ipfs node serving the pack's contents (to IPFS and/or HTTP).
    bag      create BagIt spec-compliant bag from a pack.
    car      create a `.car` certified archive from a pack.
```

### Usage Example

```
> pwd
/home/jbenet/myPack

> ls
someJSON.json
someXML.xml
moreData/

> ipfs-pack make
> ipfs-pack make -v
wrote PackManifest

> ls
someJSON.json
someXML.xml
moreData/
PackManifest

> cat PackManifest
QmVP2aaAWFe21QjUujMw5hwYRKD1eGx3yYWEBbMtuxpqXs <fmtstr> moreData/0
QmV7eDE2WXuwQnvccsoXSzK5CQGXdFfay1LSadZCwyfbDV <fmtstr> moreData/1
QmaMY7h9pmTcA5w9S2dsQT5eGLEQ1CwYQ32HwMTXAev5gQ <fmtstr> moreData/2
QmQjYU5PscpCHadDbL1fDvTK4P9eXirSwD8hzJbAyrd5mf <fmtstr> moreData/3
QmRErwActoLmffucXq7HPtefBC19MjWUcj1DdBoaAnMm6p <fmtstr> moreData/4
QmeWvL929Tdhzw27CS5ZVHD73NQ9TT1xvLvCaXCgi7a9YB <fmtstr> moreData/5
QmXbzZeh44jJEUueWjFxEiLcfAfzoaKYEy1fMHygkSD3hm <fmtstr> moreData/6
QmYL17nYZrZsAhJut5v7ooD9hmz2rBotC1tqC9ZPxzCfer <fmtstr> moreData/7
QmPKkidoUYX12PyCuKzehQuhEJofUJ9PPaX2Gc2iYd4GRs <fmtstr> moreData/8
QmQAubXA3Gji5v5oaJhMbvmbGbiuwDf1u9sYsN125mcqrn <fmtstr> moreData/9
QmYbYduoHMZAUMB5mjHoJHgJ9WndrdWkTCzuQ6yHkbgqkU <fmtstr> someJSON.json
QmeWiZD5cdyiJoS3b7h87Cs9G21uQ1sLmeKrunTae9h5qG <fmtstr> someXML.xml
QmVizQ5fUceForgWogbb2m2v5RRrE8xEm8uSkbkyNB4Rdm <fmtstr> moreData
QmZ7iEGqahTHdUWGGZMUxYRXPwSM3UjBouneLcCmj9e6q6 <fmtstr> .

> ipfs-pack repo make
> ipfs-pack repo make -v
created repo at .ipfs-pack

> ls -a
./
../
.ipfs-pack/
someJSON.json
someXML.xml
moreData/
PackManifest
```

### `ipfs-pack make` create (or update) a pack manifest

This command creates (or updates) the pack's manifest file. It writes (overwrites) the `PackManifest` file.

```
ipfs-pack make
# wrote PackManifest
```

### `ipfs-pack verify` checks whether a pack matches its manifest

This command checks whether a pack matches its `PackManifest`.

```
# errors when there is no manifest
> random-files foo
> cd foo
> ipfs-pack verify
error: no PackManifest found

# succeeds when manifest and pack match
> ipfs-pack make
> ipfs-pack verify

# errors when manifest and pack do not match
> echo "QmVizQ5fUceForgWogbb2m2v5RRrE8xEm8uSkbkyNB4Rdm <fmtstr> non-existent-file1" >>PackManifest
> echo "QmVizQ5fUceForgWogbb2m2v5RRrE8xEm8uSkbkyNB4Rdm <fmtstr> non-existent-file2" >>PackManifest
> touch non-manifest-file3
> ipfs-pack verify
error: in manifest, missing from pack: non-existent-file1
error: in manifest, missing from pack: non-existent-file2
error: in pack, missing from manifest: non-manifest-file3
```

### `ipfs-pack repo` creates (or updates) a temporary ipfs object repo

This command creates (or updates) a temporary ipfs object repo (eg at `.ipfs-pack`). This repo contains some IPLD objects and positonal metadata in the pack files for the rest (filestore). This way the only IPLD objects stored are intermediate nodes that are not the raw data leaves. The leaves' data is kept in the original files, and referenced according to the filestore tooling.

Notes:

- This repo can be considered a cache, to be destroyed and recreated as needed.
- `<filestore-descriptor>` is the file position metadata necessary to reconstruct an entire IPLD object from data in the pack.
- Intermediate ipld objects (eg intermediate objects in a file, which are not raw data nodes) may need to be stored in the db.

```
> ipfs-pack repo --help
USAGE
    ipfs-pack repo <subcommand> <arguments>

SUBCOMMANDS
    regen   regenerate ipfs-pack repo for this pack
    ls     lists all cids in the pack repo
    gc     garbage collect the pack repo (pinset = PackManifest)
    rm     removes the pack's ipfs repo
```


### `ipfs-pack serve` starts an ipfs node serving the pack's contents (to IPFS and/or HTTP).

This command starts an ipfs node serving the pack's contents (to IPFS and/or HTTP). This command MAY require a full go-ipfs installation to exist. It MAY be a standalone binary (`ipfs-pack-serve`). It MUST use an ephemeral node or a one-off node whose id would be stored locally, in the pack, at `<pack-root>/.ipfs-pack/repo`

```
> ipfs-pack serve --http
Serving pack at /ip4/0.0.0.0/tcp/1234/http - http://127.0.0.1:1234

> ipfs-pack serve --ipfs
Serving pack at /ip4/0.0.0.0/tcp/1234/ipfs/QmPVUA4rJgckcf1ifrZF5KvwV1Uib5SGjJ7Z5BskEpTaSE
```

### `ipfs-pack bag` convert to and from BagIt (spec-compliant) bags.

This command converts between BagIt (spec-compliant) bags, a commonly used [archiving format](https://tools.ietf.org/html/draft-kunze-bagit-06#section-2.1.3) very similar to `ipfs-pack`. It works like this:


```
> ipfs-pack bag --help
USAGE
  ipfs-pack-bag <src-pack> <dst-bag>
  ipfs-pack-bag <src-bag> <dst-pack>

# convert from pack to bag
> ipfs-pack bag path/to/mypack path/to/mybag

# convert from bag to pack
> ipfs-pack bag path/to/mybag path/to/mypack
```

### `ipfs-pack car` convert to and from a car (certified archive).

This command converts between packs and cars (certified archives). It works like this:


```
> ipfs-pack car --help
USAGE
  ipfs-pack-car <src-pack> <dst-car>
  ipfs-pack-car <src-car> <dst-pack>

# convert from pack to car
> ipfs-pack car path/to/mypack path/to/mycar.car

# convert from car to pack
> ipfs-pack car path/to/mycar.car path/to/mypack
```

## filestore in the repo

Part of the point of ipfs-pack is to avoid copying data, and therefore it benefits tremendously from the filestore repo datastore.

Maybe the `ipfs repo filestore` abstractions can leverage `ipfs-packs` to understand what is being tracked in a given directory, particularly if those packs have up-to-date local dbs of all their objects.


## Test Cases

WIP

## References

- [BagIt] The BagIt File Packaging Format, October 2016, <https://tools.ietf.org/html/draft-kunze-bagit-14>
