# Nebula-importer

![test](https://github.com/vesoft-inc/nebula-importer/actions?workflow=test)

## 介绍

Nebula Importer 是一款 [Nebula Graph](https://github.com/vesoft-inc/nebula) 的CSV 文件导入工具, 其读取本地的 csv 文件，然后写入到 Nebula Graph 图数据库中。

在使用 Nebula Importer 之前，首先需要部署 Nebula Graph 的服务，并且在其中创建好对应的 `space`, `tag` 和 `edge` 元数据信息。目前有两种部署方式：

1. [nebula-docker-compose](https://github.com/vesoft-inc/nebula-docker-compose "nebula-docker-compose")
2. [rpm 包安装](https://github.com/vesoft-inc/nebula/tree/master/docs/manual-EN/3.build-develop-and-administration/3.deploy-and-administrations/deployment)

> 如果想在本地快速试用 Nebula Graph，推荐使用 `docker-compose` 在本地部署

Nebula Importer 通过 YAML 配置文件来描述要导入的文件信息、Nebula Graph 的 server 信息等。下面我们就来描述配置文件中的每一项的含义。

## 配置文件

[这里](example/)有一个配置文件的参考样例和对应的数据文件格式。对应其中的每一部分的含义我们接下来逐一解释。

### 描述

```yaml
version: v1rc1
description: example
```

#### `version`

*必填*。表示配置文件的版本，默认值为 `v1rc1`。

#### `description`

*可选*。对当前配置文件的描述信息。

### `clientSettings`

```yaml
clientSettings:
  concurrency: 4
  channelBufferSize: 128
  space: test
  connection:
    user: user
    password: password
    address: 127.0.0.1:3699
```

#### `clientSettings.concurrency`

*可选*。表示 Nebula Graph Client 的并发度，即同 Nebula Graph Server 的连接数，默认为 10。

#### `clientSettings.channelBufferSize`

*可选*。表示每个 Nebula Graph Client 对应的 channel 的buffer 大小，适当的 buffer 可以缓解繁忙的 client 阻塞文件 Reader 的情况，提高并发度。

#### `clientSettings.space`

*必填*。指定所有的数据文件将要导入到哪个 `space`。

#### `clientSettings.connection`

*必填*。配置 Nebula Graph Server 的 `user`，`password` 和 `address` 信息。

### 文件

#### 日志

```yaml
logPath: ./err/test.log
```

##### `logPath`

*可选*。指定导入过程中的错误等日志信息输出的文件路径，默认输出到 `/tmp/nebula-importer.log` 中。

#### 数据

```yaml
files:
  - path: ./student.csv
    failDataPath: ./err/student.csv
    batchSize: 2
    type: csv
    csv:
      withHeader: false
      withLabel: false
```

##### `path`

*必填*。指定数据文件的存放路径，如果使用相对路径，则会去找当前配置文件的目录加上 `path`。

##### `failDataPath`

*必填*。指定插入失败的数据输出的文件，以便处理错误时，只需再次插入上面文件的数据即可。

##### `batchSize`

*可选*。批量插入的数据条数，默认 128。

##### `type` & `csv`

*必填*。指定文件的类型，目前只支持 CSV 文件导入。在 CSV 文件中可以指定是否含有头和插入和删除的标记。

- `withHeader`: 默认是 false，头的格式在后面描述。
- `withLabel`: 默认是 false，label 的格式也在后面描述。

##### `schema`

*必填*。描述当前数据文件的元数据信息。`schema.type` 只有两种值：`vertex` 和 `edge`。

- 当指定 `type: vertex` 时，需要在 `vertex` 字段中继续描述，
- 当指定 `type: edge` 时，需要在 `edge` 字段中继续描述。

###### `schema.vertex`

```yaml
    schema:
      type: vertex
      vertex:
        tags:
          - name: student
            props:
              - name: name
                type: string
              - name: age
                type: int
              - name: gender
                type: string
```

*必填*。描述插入顶点的 schema 信息，比如 tags。由于一个 VERTEX 可以含有多个 TAG，所以不同的 TAG 在 `schema.vertex.tags` 数组中给出。

对于每一个 TAG，有以下两个属性:

- `name`：TAG 的名字
- `props`：TAG 的属性字段，每个属性字段又由如下两个字段构成：
  - `name`: 属性名字，同 Nebula Graph 中创建的 TAG 中的属性名字一致。
  - `type`: 属性类型，目前支持 `bool`，`int`，`float`，`double`，`timestamp`，`string` 几种类型。

> 注意: 上述props 中的属性描述*顺序*必须同数据文件中的对应数据排列顺序一致。

###### `schema.edge`

```yaml
    schema:
      type: edge
      edge:
        name: choose
        withRanking: false
        props:
          - name: grade
            type: int
```

*必填*。描述插入边的 schema 信息。含有如下三个字段：

- `name`: 边的名字，同 Nebula Graph 中创建的 edge 名字一致。
- `withRanking`: 指定该边是否又 `rank` 值，用来区分同类型的 edge 的不同边。
- `props` 描述同上述顶点，同样需要注意跟数据文件中的排列顺序。

## Header

在 CSV 文件中，可以在第一行指定每一列的名称即数据类型。

### 没有header 的数据格式

如果在上述配置中的 `files.withHeader` 配置为 `false`，那么 CSV 文件中

#### Vertex
#### Edge

### Configuration Properties
## CSV Data Example

There will be two csv data formats supported in the future. But now please use the first format which has no header line in your csv data file.

### Without Header Line

#### Vertex

In vertex csv data file, first column could be a label(+/-) or the vid. Vertex VID column is the first column if the label option `csv.withLabel` configured `false`.
Then property values are behind VID and the order of these values must be same as `props` in configuration.

```csv
1,2,this is a property string
2,4,yet another property string
```

with label:

- `+`: Insert
- `-`: Delete

In labeled `-` row, only need the vid which you want to delete.

```csv
+,1,2,this is a property string
-,1
+,2,4,yet anthor property string
```

#### Edge

Edge csv data file format is like the vertex description. But difference with above vertex vid is source vertex vid, destination vertex vid and edge ranking.

Without label column, `src_vid`, `dst_vid` and `ranking` always are first three columns in csv data file.

```csv
1,2,0,first property value
1,3,2,prop value
```

Ranking column is not required, you must not give it if you don't need it.

```csv
1,2,first property value
1,3,prop value
```

with label:

```csv
+,1,2,0,first property value
+,1,3,2,prop value
```

### With Header Line

#### Format

`<type.field_name>:<field_type>`, `field_type` default type is `string`.

#### Edge

```csv
:SRC_VID,:DST_VID,:RANK,prop1,prop2
...
```

`:SRC_VID` and `:DST_VID` represent edge source and destination vertex id. `:RANK` column is value of edge ranking.

#### Vertex

```csv
:VID,tag1.prop1:string,tag2.prop2:int,tag1.prop3:string,tag2.prop4:int
...
```

`:VID` column represent the global unique vertex id.

#### Skipping columns

```csv
:VID,name,:IGNORE,age:int
```

## Usage

### From Sources

This tool depends on golang 1.13, so make sure you have install `go` first.

Use `git` to clone this project to your local directory and execute the `cmd/importer.go` with `config` parameter.

``` shell
$ git clone https://github.com/vesoft-inc/nebula-importer.git
$ cd nebula-importer/cmd
$ go run importer.go --config /path/to/yaml/config/file
```

### Docker

With docker, we can easily to import our local data to nebula without `golang` runtime environment.

```shell
$ docker run --rm -ti \
    --network=host \
    -v {your-config-file}:{your-config-file} \
    -v {your-csv-data-dir}:{your-csv-data-dir} \
    vesoft/nebula-importer
    --config {your-config-file}
```

### Log

All logs info will output to your `logPath` file in configuration.

## TODO

- [X] Summary statistics of response
- [X] Write error log and data
- [X] Configure file
- [X] Concurrent request to Graph server
- [ ] Create space and tag/edge automatically
- [ ] Configure retry option for Nebula client
- [X] Support edge rank
- [X] Support label for add/delete(+/-) in first column
- [ ] Support column header in first line
- [X] Support vid partition
- [X] Support multi-tags insertion in vertex
- [X] Provide docker image and usage
- [ ] Make header adapt to props order defined in schema of configure file
- [X] Handle string column in nice way
- [ ] Update concurrency and batch size online
- [ ] Count duplicate vids
- [ ] Support VID generation automatically
- [ ] Output logs to file
