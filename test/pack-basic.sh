#!/bin/bash

test_pack() {
	exp="$1"
	data="$2"
	rm -rf packtest
	mkdir -p packtest
	mv "$2" packtest/
	cd packtest
	ipfs-pack make 

	if ! grep "$exp" PackManifest; then
		echo "pack hash not as expected"
		echo "got " $(tail -n1 PackManifest)
		exit 1
	fi

	if ! ipfs-pack verify > verify_out; then
		echo "pack verify failed"
		cat verify_out
		exit 1
	fi

	echo "pack verify succeeded"
}

random-files -seed=42 -files=6 -depth=3 -dirs=5 stuff
test_pack QmNeY1rrxVCBm1XHBeByHZrBEyfk4eqC1ckY9c6uVaLN8r stuff

echo "foo" > afile
test_pack QmaSFTnDx8iQJpih2tsJCwYmhnNpAggt2oSc49UZongcPS afile
