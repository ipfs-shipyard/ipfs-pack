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

# Quote arguments for sh eval
shellquote() {
	_space=''
	for _arg
	do
		# On Mac OS, sed adds a newline character.
		# With a printf wrapper the extra newline is removed.
		printf "$_space'%s'" "$(printf "%s" "$_arg" | sed -e "s/'/'\\\\''/g;")"
		_space=' '
	done
	printf '\n'
}

# Echo the args, run the cmd, and then also fail,
# making sure a test case fails.
test_fsh() {
    echo "> $@"
    eval $(shellquote "$@")
    echo ""
    false
}

# Same as sharness' test_cmp but using test_fsh (to see the output).
# We have to do it twice, so the first diff output doesn't show unless it's
# broken.
test_cmp() {
	diff -q "$@" >/dev/null || test_fsh diff -u "$@"
}

# Same as test_cmp above, but we sort files before comparing them.
test_sort_cmp() {
	sort "$1" >"$1_sorted" &&
	sort "$2" >"$2_sorted" &&
	test_cmp "$1_sorted" "$2_sorted"
}

