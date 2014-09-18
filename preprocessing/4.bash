#!/bin/bash

IN=$1
OUT=$2

cat $IN | go run select_types.go >> $OUT
