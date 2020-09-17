# Build Go environment

## Download the installation package

- https://studygolang.com/dl

## Unzip the package and move it to /usr/local/go

```bash
$ mv golang-1.13 /usr/local/go
```

## Configure environment variables

```bash
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export GOPROXY=https://goproxy.cn
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

Add the preceding configurations to the `~/.bashrc` file. Run the `source ~/.bashrc` command to take effect.

## Verify your installation

```bash
$ go version
```

## Compile nebula-importer

Go to the nebula-importer project directory and run the following commands:

```bash
$ cd nebula-importer/cmd
$ go build -mod vendor -o nebula-importer
$ ./nebula-importer --help
```
