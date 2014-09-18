#!/bin/bash
FILES=/data/freebase/a/a*
OUT=all_types.nt.gz
for f in $FILES
do
	echo $f
	cat $f >> $OUT
done
