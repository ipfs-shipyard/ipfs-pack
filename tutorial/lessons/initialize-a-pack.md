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
```

## Step 3: Build a Local IPFS Object Store inside the Pack

This will build an IPFS repository in the root of the pack

```
$ ipfs-pack repo --help
```

```
$ ipfs-pack repo make
```
