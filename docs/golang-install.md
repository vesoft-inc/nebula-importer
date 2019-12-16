Golang 环境搭建
===========

## 下载安装包

- https://studygolang.com/dl

## 解压并移动到 /usr/local/go

```bash
$ mv golang-1.13 /usr/local/go
```

## 配置环境变量

```bash
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export GOPROXY=https://goproxy.cn
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

将上述配置加到 `~/.bashrc` 文件中，并通过 `source ~/.bashrc` 使其生效。

## 检验是否安装成功

```bash
$ go version
```

## 编译 nebula-importer

首先进入 nebula-importer 的项目目录。然后执行如下的命令：

```bash
$ cd nebula-importer/cmd
$ go build -mod vendor -o nebula-importer
$ ./nebula-importer --help
```
