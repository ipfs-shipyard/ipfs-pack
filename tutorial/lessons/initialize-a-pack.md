# Lesson: Initialize an IPFS Pack

## Goals

After doing this Lesson you will be able to

* Initialize a directory on your machine as an IPFS Pack
* Serve the contents of an IPFS Pack over the IPFS network
* Explain the contents of the PackManifest file in the root of your IPFS Pack

## Steps

### Step 1: Create a Pack
This will create a PackManifest file in the directory you specified. The PackManifest represents that directory and all the files & directories within it as a [pack tree](../spec.md#terms).
The directory you specified, where the PackManifest file is generated, is referred to as the [pack root](../spec.md#terms)  

First `cd` into the root directory of a dataset that you want to track.
```
$ cd {path to your directory}
```

While in the root directory of the dataset, run `ipfs-pack make`.
```
$ ipfs-pack make
wrote PackManifest
```

Contratulations. Now you've built a Pack for your dataset. It created a `PackManifest` file, which lists the directory's contents and their hashes, and an `.ipfs-pack` directory which contains the IPFS repo (aka. object store) that represents your dataset at hashed, content-addressed chunks that are ready to be served to the network. Remember, unlike normal IPFS repositories which actually contain a copy of your data, the `.ipfs-pack` repo _only contains references to the data_. They're referenced by paths relative to the root of your pack.

### Step 2: Inspect the PackManifest

Use a text editor to look at the contents of the `PackManifest` file.

```
$ vi PackManifest
```

It has a bunch of lines that look like this:
```
zb2rhmyBBjTBw1q9TQQv3YV69Gc2tHccTv7egjNNVXuu8YPpw       f0000120001     ./tutorial/lessons/initialize-a-pack.md
```

The format of each line is:
`<content hash> <format string> <relative path>`

The `<content hash>` is the hash of the file located at `<relative path>`. The `<format string>` provides a [future-proofed](https://www.youtube.com/watch?v=soUG72j7kB0) record of how the content hash was generated and how the content was added to the `.ipfs-pack` repo.

### Step 3: Verify the contents of the directory against the PactManifest

At any point you can use the PackManifest to verify that the contents of your pack have not changed. To do this, run `ipfs-pack verify`

```
$ ipfs-pack verify
Pack verified successfully!
```

If any of the files in your pack have changed, the output will tell you which files have changed and conclude with the message `Pack verify found some corruption.` For example, if my pack contaiined a file called `styles/site.css` and I modified it after building the pack, I would get a message like:

```
Checksum mismatch on ./styles/site.css. (zb2rhof9xknpBt36jvWRPVADLfsk2zhL7y5dLUSRRNfMuTGnF)
Pack verify found some corruption.
```

For more info about using this command, read the helptext on the command line.
```
$ ipfs-pack verify --help
```

### Step 4: Update the PackManifest

ipfs-pack is not like git, which keeps previous versions of your files when you commit new changes. On the contrary, the current version of ipfs-pack is specifically designed **not to duplicate** any of your content. This is so you can add any amount of data, possibly hundreds of Terabytes, without taking up extra storage. For more info about this, read the lesson on [understanding ipfs-pack](understanding-ipfs-pack.md).

If you change any of the files in your pack, you can update the PackManifest by running `ipfs-pack make` again. This will update the PackManifest to accurately represent the current contents of the pack tree. , Note that this will change the root hash of the pack and if you remove or modify any of the dataset contents the old information will no longer be available from your ipfs-pack node.

### Step 5: Rebuild the Local IPFS Object Store inside your Pack

When you ran `ipfs-pack make` it built an IPFS repository in the root of your pack called `.ipfs-pack`. If you want to regenerate that repository, run `ipfs-pack repo regen`

```
$ ipfs-pack repo regen
```

To see the other commands you can run against the pack repo, run

```
$ ipfs-pack repo --help
```

## Next Steps

Now you're ready to [serve the pack contents over ipfs](serve-pack-contents.md).
