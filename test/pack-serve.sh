#!/bin/bash

rm -rf packtest
mkdir -p packtest
cd packtest
random-files -seed=42 -files=6 -depth=3 -dirs=5 stuff
ipfs-pack make 

ipfs-pack serve > serve_out &
echo $! > serve-pid
sleep 1

grep -A1 "Peer ID" serve_out >addrinfo
PEERID=$(cat addrinfo | head -n1 | awk '{ print $3 }')
ADDR=$(cat addrinfo | tail -n1 | awk '{ print $1 }')
echo peerid $PEERID
echo addr $ADDR

rm -rf ../ipfs
export IPFS_PATH=../ipfs
ipfs init
ipfs bootstrap rm --all
ipfs config --json Discovery.MDNS.Enabled false
ipfs daemon &
echo $! > ipfs-pid
sleep 1
ipfs swarm connect $ADDR/ipfs/$PEERID

HASH=$(tail -n1 PackManifest | awk '{ print $1 }')

ipfs refs --timeout=10s -r $HASH

kill $(cat serve-pid)
kill $(cat ipfs-pid)
