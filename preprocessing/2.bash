#!/bin/bash

IN=$1
OUT=$2

cat $IN | python select_aliases.py >> $OUT
