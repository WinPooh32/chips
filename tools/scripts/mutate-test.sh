#!/bin/bash

# Exec script for go-mutesting using gotestsum as the test runner.
# Replaces the original file with the mutation, runs gotestsum,
# and reports whether the mutation was killed.

if [ -z ${MUTATE_CHANGED+x} ]; then echo "MUTATE_CHANGED is not set"; exit 1; fi
if [ -z ${MUTATE_ORIGINAL+x} ]; then echo "MUTATE_ORIGINAL is not set"; exit 1; fi
if [ -z ${MUTATE_PACKAGE+x} ]; then echo "MUTATE_PACKAGE is not set"; exit 1; fi

function clean_up {
	if [ -f $MUTATE_ORIGINAL.tmp ];
	then
		mv $MUTATE_ORIGINAL.tmp $MUTATE_ORIGINAL
	fi
}

function sig_handler {
	clean_up

	exit $GOMUTESTING_RESULT
}
trap sig_handler SIGHUP SIGINT SIGTERM

export GOMUTESTING_DIFF=$(diff -u $MUTATE_ORIGINAL $MUTATE_CHANGED)

SKIP_PATTERN="^-.*(\.log\.\w+Context|fmt\.F?print|fmt\.Errorf|return err|return nil, err|return.*fmt\.Errorf)"

if echo "$GOMUTESTING_DIFF" | grep -q -i -E "$SKIP_PATTERN"
then # skip
	exit 2
fi

mv $MUTATE_ORIGINAL $MUTATE_ORIGINAL.tmp
cp $MUTATE_CHANGED $MUTATE_ORIGINAL

export MUTATE_TIMEOUT=${MUTATE_TIMEOUT:-10}

if [ -n "$TEST_RECURSIVE" ]; then
	TEST_RECURSIVE="/..."
fi

GOMUTESTING_TEST=$(go tool -modfile=misc/gotestsum-go.mod gotestsum --max-fails=1 -- -short -timeout $(printf '%ds' $MUTATE_TIMEOUT) -failfast $MUTATE_PACKAGE$TEST_RECURSIVE 2>&1)
export GOMUTESTING_RESULT=$?

if [ "$MUTATE_DEBUG" = true ] ; then
	echo "$GOMUTESTING_TEST"
fi

clean_up

case $GOMUTESTING_RESULT in
0) # tests passed -> FAIL
	echo "$GOMUTESTING_DIFF"

	exit 1
	;;
1) # tests failed -> PASS
	if [ "$MUTATE_DEBUG" = true ] ; then
		echo "$GOMUTESTING_DIFF"
	fi

	exit 0
	;;
2) # did not compile -> SKIP
	if [ "$MUTATE_VERBOSE" = true ] ; then
		echo "Mutation did not compile"
	fi

	if [ "$MUTATE_DEBUG" = true ] ; then
		echo "$GOMUTESTING_DIFF"
	fi

	exit 2
	;;
3) # tests failed ? -> PASS
	echo "$GOMUTESTING_DIFF"

	exit 0
	;;
*) # Unknown exit code -> SKIP
	echo "Unknown exit code $GOMUTESTING_RESULT"
	echo "$GOMUTESTING_DIFF"

	exit $GOMUTESTING_RESULT
	;;
esac
