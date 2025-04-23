#!/bin/bash

set -e
set -x

function trap_handler {
  MYSELF="$0"   # equals to my script name
  LASTLINE="$1" # argument 1: last line of error occurence
  LASTERR="$2"  # argument 2: error code of last command
  echo "Error: line ${LASTLINE} - exit status of last command: ${LASTERR}"
  exit $2
}
trap 'trap_handler ${LINENO} ${$?}' ERR

echo "Checking dependencies..."
which protoc
which protoc-gen-go


echo "Building protobuf code..."
DIR=`pwd`
echo "DIR $DIR"

echo "changed dir to scripts"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
echo "after DIR $DIR"

SRCDIR=$GOPATH/src

echo "DIR $DIR"
echo "SRCDIR $SRCDIR"
find $DIR/proto -name '*.pb.go' -exec rm {} \;
find $DIR/proto -name '*.proto' -exec echo {} \;
# find $DIR/proto -name '*.proto' -exec protoc --proto_path=$SRCDIR --micro_out=${MOD}:${SRCDIR} --go_out=${MOD}:${SRCDIR} {} \;


echo "Complete"
