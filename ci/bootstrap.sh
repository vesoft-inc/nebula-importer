#!/bin/sh

set -e

addr=$1
port=$2

export GOPATH=/usr/local/nebula/
export GO111MODULE=on

# build nebula-console
mkdir -p nebulaconsolebuild
cd nebulaconsolebuild
  wget "https://github.com/vesoft-inc/nebula-console/archive/master.zip" -O nebula-console.zip
  unzip ./nebula-console.zip -d ./
  cd nebula-console-master
    go build -o ../../nebula-console
  cd ..
cd ..
rm -rf nebulaconsolebuild

cd /usr/local/nebula/importer/cmd
go build -o ../../nebula-importer
cd /usr/local/nebula

until echo "quit" | /usr/local/nebula/nebula-console -u root -p password --addr=$addr --port=$port &> /dev/null; do
  echo "nebula graph is unavailable - sleeping"
  sleep 2
done

echo "nebula graph is up - executing command"
./nebula-importer --config ./importer/examples/v2/example.yaml
./nebula-importer --config ./importer/examples/v1/example.yaml
