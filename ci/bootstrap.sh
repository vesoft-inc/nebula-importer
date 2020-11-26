#!/bin/sh

set -e

addr=$1
port=$2

export GOPATH=/usr/local/nebula/
export GO111MODULE=on

# build nebula-console
wget "https://github.com/vesoft-inc/nebula-console/archive/master.zip" -O nebula-console.zip
unzip nebula-console.zip -d .
mv nebula-console-* nebula-console
cd nebula-console
go build -o nebula-console

cd /usr/local/nebula/importer/cmd
go build -o ../../nebula-importer
cd /usr/local/nebula

until echo "quit" | /usr/local/nebula/nebula-console/nebula-console -u user -p password --addr=$addr --port=$port &> /dev/null; do
  echo "nebula graph is unavailable - sleeping"
  sleep 2
done

echo "nebula graph is up - executing command"
./nebula-importer --config ./importer/examples/example.yaml
