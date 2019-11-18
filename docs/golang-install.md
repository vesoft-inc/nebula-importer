Golang 环境搭建
===========

## 下载安装包

## 移动到 /usr/local/go

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

## 检验是否安装成功

```bash
$ go version
```
