#!/bin/bash

IN=$1

cat $IN | python vocabulary.py > vocabulary.txt
cat $IN | python build_training_set.py > training_set.csv
