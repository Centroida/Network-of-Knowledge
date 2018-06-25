#!/bin/bash

set x

haveIPFS=$(which ipfs)

if [ -z $haveIPFS ]; then
	source ./install-ipfs.sh
fi

sudo systemctl restart ipfs &
