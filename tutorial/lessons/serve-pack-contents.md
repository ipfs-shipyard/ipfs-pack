# Lesson: Serve the Contents of the Pack to the Network

## Goals

After doing this Lesson you will be able to

* Serve the contents of an IPFS pack to the IPFS network

## Steps

### Step 1: Start the IPFS node

ipfs-pack lets you serve the contents of your pack directly to the ipfs network. When you created the pack with the `ipfs-pack make` command, it set up everything you need. Now all you need to do is start the node so it can connect to the network and serve the content.  To do this, run

```
$ ipfs-pack serve
```

After starting the ipfs node with `ipfs-pack serve`, you will see some info about the node printed on the console. It will look like:

```
verified pack, starting server...
Serving data in this pack...
Peer ID: QmVbXV7mQ5Fs3tYY2Euek5YdkkzcRafUg8qGWvFdgaBMuo
    /ip4/127.0.0.1/tcp/58162
    /ip4/1.2.3.4/tcp/58162
Pack root is QmRguPt6jHmVMzu1NM8wQmpoymM9UeqDJGXdQyU3GhiPy4
Shared:      0 blocks, 0 B total data uploaded
```

### Step 2: Share the Pack Root Hash

Now other nodes on the network can access the dataset, but they need to know the hashes of the dataset's content. The easiest way to give them access to that info is by giving them the pack root hash. That hash is in the information that was printed on the command line when you ran `ipfs-pack serve`. In the example output above, it's the second to last line, which reads:

```
Pack root is QmRguPt6jHmVMzu1NM8wQmpoymM9UeqDJGXdQyU3GhiPy4
```

If you give that hash, which usually begins with `Qm`, to anyone else they can use it to request the dataset from your node.

### Step 3: Try it out

If you have a copy of regular ipfs (not ipfs-pack) installed, you can confirm that your dataset is available on the network by running a second ipfs node on your machine and using it to read the data. For example, you can list the contents of your pack's root directory `ipfs ls`

First, start a second, regular ipfs node. This will let your ipfs node retrieve the data from your pack repo using the ipfs protocol.

```
$ ipfs daemon
```

Now list the contents of your pack's root directory using `ipfs ls`
```
$ ipfs ls YOUR_ROOT_HASH
```

## Next Steps

You're all set! Go to IRC or the IPFS forums to tell us about the data you're serving with ipfs-pack. For info about how to connect with the community, visit https://github.com/ipfs/community
