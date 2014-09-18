#!/bin/bash

# Convert input XML file to utf-8 and extract sentences.

FILE=$1
OUT=$2

iconv -f ISO-8859-1 -t utf-8 $FILE | sed s/ISO-8859-1/utf-8/ | \
	go run unescape_html.go | go run extract_sentences.go >> $OUT
