#!/bin/bash
FILES=/data/parlparse/documents/standing/standing201*
for f in $FILES
do
	echo "Handling $f"
	/bin/bash /data/politics/1.bash $f speech.csv
done

