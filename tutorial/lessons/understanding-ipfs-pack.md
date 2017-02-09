Lesson: Understanding IPFS Pack
===

## Goals

After doing this Lesson you will be able to
* Understand the use cases that motivated the creation of ipfs-pack
* Compare and contrast the "copy-on-add" and "nocopy" approaches to adding content to ipfs
* Decide whether IPFS Pack is the right tool for your needs
* Explain terms from the ipfs-pack Specification, such as "pack tree", "pack root" and "PackManifest"


## Steps

### The Motivation

We designed ipfs-pack to support situations where you have a large dataset that you do not intend to change.

This is not like git, which keeps previous versions of your files when you commit new changes. On the contrary, the current version of ipfs-pack is specifically designed **not to duplicate** any of your content. This is so you can add any amount of data, possibly hundreds of Terabytes, without taking up much extra storage.

### Is ipfs-pack the Right Tool for You?

copy-on-add vs "nocopy"

Over time, ipfs-pack might evolve to gracefully handle situations where the registered content _can_ change.

Not appropriate for accumulating a version history of a dataset that changes over time. For that you should use regular ipfs, adding the content to your ipfs repository with `ipfs add` whenever it changes. That way, the copy of your data that accumulates in the ipfs repository will contain all versions of the data and will attempt to store those versions in a compact way that minimizes duplication.

### Don't use ipfs-pack for...
* **Don't use for** Sharing files that you intend to change frequently
* **Don't use for** Tracking version histories of content

For those situations, use regular ipfs, not ipfs-pack.

#### Good uses for ipfs-pack
* Sharing archival copies of content which you don't intend to change
* Sharing large volumes of data that you can't afford to duplicate

**Example: Serving and Archival Copy** If you have an archival copy of a dataset that you actively plan to preserve in its current structure, you can use ipfs-pack to serve that archival copy directly to the ipfs network. That way, it does double-duty. It's both an archival copy and and a network-available seed of the dataset.

**Example: Serving a Dataset that's Too Big to Duplicate** If you want to serve a dataset on IPFS but don't have enough storage space available to store a second copy of the data in your IPFS repository, you can use ipfs-pack to serve the dataset directly from its current location.

For both of these examples, if a situation comes up where you have to change or rearrange the dataset, you can always rebuild the pack, but note that this will change the root hash of the pack and if you remove or modify any of the dataset contents they will cease to be available through your ipfs-pack node -- _because_ ipfs-pack doesn't copy the data into your IPFS repo, there's no backup copy. With the ipfs-pack approach, the files in your pack are the only copies of the data.

### The ipfs-pack Specification

Read [the ipfs-pack Spec](../../spec.md) to learn more about the design of ipfs-pack. Especially pay attention to the [terms](../../spec.md#terms) section to learn what we mean by "pack tree", "pack root", "pack repo" and "PackManifest".

## Next Steps

Now that you've read about ipfs-pack, either [install it](install-ipfs-pack.md) or proceed to [initialize a pack](initialize-a-pack.md).
