#!/bin/bash

FILES=/data/freebase/label/*.nt.gz
for f in $FILES
do
	echo "Handling $f"
	/bin/bash /data/politics/3.bash $f entities.txt
done

