# Lesson: Initialize an IPFS Pack

## Goals

After doing this Lesson you will be able to

* Initialize a directory on your machine as an IPFS Pack
* Serve the contents of an IPFS Pack over the IPFS network
* Explain the contents of the PackManifest file in the root of your IPFS Pack

## Steps

## Step 1: Create a Pack
This will create a PackManifest file in the directory you specified. The PackManifest represents that directory and all the files & directories within it as a [pack tree](../spec.md#terms).
The directory you specified, where the PackManifest file is generated, is referred to as the [pack root](../spec.md#terms)  

```
$ cd {path to your directory}
```

```
$ ipfs-pack make
wrote PackManifest
```

## Step 2: Inspect the PackManifest

```
$ vi PackManifest
```

Example line:
```
zb2rhmyBBjTBw1q9TQQv3YV69Gc2tHccTv7egjNNVXuu8YPpw       f0000120001     ./tutorial/lessons/initialize-a-pack.md
```

`<content hash> <format string> <relative path>`

## Step 3: Verify the contents of the directory against the PactManifest

```
$ ipfs-pack verify --help
```

```
$ ipfs-pack verify
Pack verified successfully!
```

If any of the files in your pack have changed, the output will tell you which files have changed and conclude with the message `Pack verify found some corruption.` For example, if my pack contaiined a file called `styles/site.css` and I modified it after building the pack, I would get a message like:

```
Checksum mismatch on ./styles/site.css. (zb2rhof9xknpBt36jvWRPVADLfsk2zhL7y5dLUSRRNfMuTGnF)
Pack verify found some corruption.
```

## Step 3: Update the PackManifest

If you change any of the files in your pack, you can update the PackManifest by running `ipfs-pack make` again. This will update the PackManifest to accurately represent the current contents of the pack tree.

This is not like git, which keeps previous versions of your files when you commit new changes. On the contrary, the current version of ipfs-pack is specifically designed **not to duplicate** any of your content. This is so you can add any amount of data, possibly hundreds of Terabytes, without taking up extra storage.

## Step 4: Build a Local IPFS Object Store inside the Pack

This will build an IPFS repository in the root of your pack

```
$ ipfs-pack repo --help
```

```
$ ipfs-pack repo make
```
