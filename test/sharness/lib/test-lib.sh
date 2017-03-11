# Test framework for ipfs-pack
#
# Copyright (c) 2017 Christian Couder
# MIT Licensed; see the LICENSE file in this repository.
#
# We are using sharness (https://github.com/chriscool/sharness)
# which was extracted from the Git test framework.

# Add current directory to path, for ipfs-pack.
PATH=$(pwd)/bin:${PATH}

# Set sharness verbosity. we set the env var directly as
# it's too late to pass in --verbose, and --verbose is harder
# to pass through in some cases.
test "$TEST_VERBOSE" = 1 && verbose=t

# assert the `ipfs-pack` we're using is the right one.
if test `which ipfs-pack` != $(pwd)/bin/ipfs-pack; then
	echo >&2 "Cannot find the tests' local ipfs-pack tool."
	echo >&2 "Please check ipfs-pack installation."
	exit 1
fi

SHARNESS_LIB="lib/sharness/sharness.sh"

. "$SHARNESS_LIB" || {
	echo >&2 "Cannot source: $SHARNESS_LIB"
	echo >&2 "Please check Sharness installation."
	exit 1
}

# Please put ipfs-pack specific shell functions below

if test "$TEST_VERBOSE" = 1; then
	echo '# TEST_VERBOSE='"$TEST_VERBOSE"
fi

test_fsh() {
	echo "> $@"
	eval "$@"
	echo ""
	false
}

