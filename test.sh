#!/bin/bash

FOLDERS="gnlp counter frozencounter smoothing features minimizer"

for folder in $FOLDERS; do
	pushd $folder > /dev/null
	if [ -e *_test.go ]; then
		echo "Running tests for $folder"
		make test
	fi
	popd > /dev/null
done
