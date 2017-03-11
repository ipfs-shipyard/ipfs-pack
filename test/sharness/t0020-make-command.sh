#!/bin/sh
#
# Copyright (c) 2017 Christian Couder
# MIT Licensed; see the LICENSE file in this repository.
#

test_description="Test 'make' command"

. lib/test-lib.sh

# We should work in a subdir, so that we can freely
# create files in the parent dir.

test_expect_success "Create sample data set" '
	mkdir workdir &&
	(
		cd workdir &&
		echo Hello >hello.txt &&
		mkdir subdir &&
		echo "Hello in subdir" >subdir/hello_in_subdir.txt
	)
'

test_expect_success "'ipfs-pack make' succeeds" '
	(
		cd workdir &&
		ipfs-pack make >../actual
	)
'

test_expect_success "'ipfs-pack make' output looks good" '
	grep "Building IPFS Pack" actual &&
	grep "wrote PackManifest" actual
'

test_expect_success "PackManifest looks good" '
	grep "subdir/hello_in_subdir.txt" workdir/PackManifest &&
	grep "hello.txt" workdir/PackManifest
'

test_expect_success "'ipfs-pack verify' succeeds" '
	(
		cd workdir &&
		ipfs-pack verify >../actual
	)
'

test_expect_success "'ipfs-pack verify' output looks good" '
	echo "pack verification succeeded" >expected &&
	test_cmp expected actual
'

test_done
