#!/bin/sh
#
# Copyright (c) 2017 Christian Couder
# MIT Licensed; see the LICENSE file in this repository.
#

test_description="Test some basic commands"

. lib/test-lib.sh

test_expect_success "current dir is writable" '
	echo "It works!" >test.txt
'

test_expect_success "'ipfs-pack --version' succeeds" '
	ipfs-pack --version >version.txt
'

test_expect_success "'ipfs-pack --version' output looks good" '
	egrep "^ipfs-pack version v[0-9]+\.[0-9]+\.[0-9]" version.txt >/dev/null ||
	test_fsh cat version.txt
'

test_expect_success "'ipfs-pack --help' and 'ipfs-pack help' succeed" '
	ipfs-pack --help >help1.txt &&
	ipfs-pack help >help2.txt
'

test_expect_success "'ipfs-pack --help' and 'ipfs-pack help' output look good" '
	egrep -i "A filesystem packing tool" help1.txt >/dev/null &&
	egrep -i "A filesystem packing tool" help2.txt >/dev/null &&
	egrep "ipfs-pack" help1.txt >/dev/null ||
	test_fsh cat help1.txt &&
	egrep "ipfs-pack" help2.txt >/dev/null ||
	test_fsh cat help2.txt
'

test_expect_success "'ipfs-pack help' output contain commands" '
	grep -A4 COMMANDS help2.txt | tail -4 | cut -d" " -f6 >commands.txt &&
	for cmd in $(cat commands.txt)
	do
		ipfs-pack help "$cmd" >"$cmd.txt" || break
		grep "ipfs-pack $cmd" "$cmd.txt" || break
	done
'

test_done
