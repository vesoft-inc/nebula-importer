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
for i in `seq 1 30`;do
  echo "Adding hosts..."
  var=`/usr/local/nebula/nebula-console -addr graphd1 -port 9669 -u root -p nebula -e 'ADD HOSTS "storaged":9779'`;
  if [[ $$? == 0 ]];then
    echo "Add hosts succeed"
    break;
  fi;
  sleep 1;
  echo "retry to add hosts.";
done

./nebula-importer --config ./importer/examples/v1/example.yaml
./nebula-importer --config ./importer/examples/v2/example.yaml
