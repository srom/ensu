#!/bin/bash

cat types.nt.gz | prune_entities.go > entities_final.txt
cat labels.nt.gz | prune_aliases.go > aliases_final.txt
