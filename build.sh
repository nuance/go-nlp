#!/bin/bash

FOLDERS="gnlp counter frozencounter smoothing features minimizer examples/naivebayes examples/maxent"

for folder in $FOLDERS
do 
	pushd $folder > /dev/null
	echo "* Building $folder *"
	make clean install
	popd > /dev/null
done
