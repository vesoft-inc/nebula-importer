<div align="center">
  <h1>Nebula Importer</h1>
  <div>
    <a href="https://github.com/vesoft-inc/nebula-importer/blob/master/README.md">EN</a>
  </div>
</div>

[![test](https://github.com/vesoft-inc/nebula-importer/workflows/test/badge.svg)](https://github.com/vesoft-inc/nebula-importer/actions?workflow=test)

## 介绍

Nebula Importer 是一款 [Nebula Graph](https://github.com/vesoft-inc/nebula) 的 CSV 文件导入工具, 其读取本地的 CSV 文件，然后写入到 Nebula Graph 图数据库中。

在使用 Nebula Importer 之前，首先需要部署 Nebula Graph 的服务，并且在其中创建好对应的 `space`, `tag` 和 `edge` 元数据信息。目前有两种部署方式：

1. [nebula-docker-compose](https://github.com/vesoft-inc/nebula-docker-compose "nebula-docker-compose")
2. [rpm 包安装](https://github.com/vesoft-inc/nebula/tree/master/docs/manual-EN/3.build-develop-and-administration/3.deploy-and-administrations/deployment)

> 如果想在本地快速试用 Nebula Graph，推荐使用 `docker-compose` 在本地部署。

## 配置文件

Nebula Importer 通过 YAML 配置文件来描述要导入的文件信息、Nebula Graph 的 server 信息等。[这里](example/)有一个配置文件的参考样例和对应的数据文件格式。接下来逐一解释各个选项的含义：

```yaml
version: v1rc1
description: example
```

### `version`

**必填**。表示配置文件的版本，默认值为 `v1rc1`。

### `description`

**可选**。对当前配置文件的描述信息。

### `clientSettings`

跟 Nebula Graph 服务端相关的配置均在该字段下配置。

```yaml
clientSettings:
  concurrency: 10
  channelBufferSize: 128
  space: test
  connection:
    user: user
    password: password
    address: 127.0.0.1:3699
```

#### `clientSettings.concurrency`

**可选**。表示 Nebula Graph Client 的并发度，即同 Nebula Graph Server 的连接数，默认为 10。

#### `clientSettings.channelBufferSize`

**可选**。表示每个 Nebula Graph Client 对应的 channel 的 buffer 大小，适当的 buffer 可以缓解繁忙的 client 阻塞文件 Reader 的情况，提高并发度。

#### `clientSettings.space`

**必填**。指定所有的数据文件将要导入到哪个 `space`。

#### `clientSettings.connection`

**必填**。配置 Nebula Graph Server 的 `user`，`password` 和 `address` 信息。

### 文件

跟日志和数据文件相关的配置跟以下两个选项有关：

- `logPath`: **可选**。指定导入过程中的错误等日志信息输出的文件路径，默认输出到 `/tmp/nebula-importer.log` 中。
- `files`: **必填**。数组类型，用来配置不同的数据文件。

#### 数据文件

一个数据文件中只能存放一种顶点或者边，不同 schema 的顶点或者边数据需要放置在不同的文件中。

```yaml
files:
  - path: ./student.csv
    failDataPath: ./err/student.csv
    batchSize: 128
    type: csv
    csv:
      withHeader: false
      withLabel: false
```

##### `path`

**必填**。指定数据文件的存放路径，如果使用相对路径，则会拼接当前配置文件的目录和 `path`。

##### `failDataPath`

**必填**。指定插入失败的数据输出的文件，以便后面补写出错数据。

##### `batchSize`

**可选**。批量插入数据的条数，默认 128。

##### `type` & `csv`

**必填**。指定文件的类型，目前只支持 CSV 文件导入。在 CSV 文件中可以指定是否含有文件头和插入、删除的标记。

- `withHeader`: 默认是 `false`，文件头的格式在后面描述。
- `withLabel`: 默认是 `false`，label 的格式也在后面描述。

##### `schema`

**必填**。描述当前数据文件的元数据信息。`schema.type` 只有两种取值：`vertex` 和 `edge`。

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

**必填**。描述插入顶点的 schema 信息，比如 tags。由于一个 VERTEX 可以含有多个 TAG，所以不同的 TAG 在 `schema.vertex.tags` 数组中给出。

对于每一个 TAG，有以下两个属性:

- `name`：TAG 的名字，
- `props`：TAG 的属性字段数组，每个属性字段又由如下两个字段构成：
  - `name`: 属性名字，同 Nebula Graph 中创建的 TAG 的属性名字一致。
  - `type`: 属性类型，目前支持 `bool`，`int`，`float`，`double`，`timestamp` 和 `string` 几种类型。

> 注意: 上述 props 中的属性描述**顺序**必须同数据文件中的对应数据排列顺序一致。

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

**必填**。描述插入边的 schema 信息。含有如下三个字段：

- `name`：边的名字，同 Nebula Graph 中创建的 edge 名字一致。
- `withRanking`：指定该边是否有 `rank` 值，用来区分同顶点同类型的不同边。
- `props`：描述同上述顶点，同样需要注意跟数据文件中列的排列顺序一致。

所有配置的选项解释见[表格](docs/configuration-reference.md)。

## CSV Header

针对 CSV 文件，除了在上述配置中描述每一列的 schema 信息，还可以在文件中的第一行指定对应的名称和数据类型格式。

### 没有header 的数据格式

如果在上述配置中的 `csv.withHeader` 配置为 `false`，那么 CSV 文件中只含有数据，对于顶点和边的数据示例如下：

#### 顶点

example 中 course 顶点的部分数据：

```csv
101,Math,3,No5
102,English,6,No11
```

上述中的第一列为顶点的 `:VID`，后面的三个属性值，分别按序对应配置文件中的 `vertex.tags.props`：course.name, course.credits 和 building.name。

#### 边

example 中 choose 边的部分数据：

```csv
200,101,5
200,102,3
```

上述中的前两列的数据分别为 `:SRC_VID` 和 `:DST_VID`，最后一列对应 choose.likeness 属性值。
如果上述边中含有 rank 值，请在第三列放置 rank 的值。

### 含有 header 的数据格式

如果配置文件中 `csv.withHeader` 设置为 `true`，那么对应的数据文件中的第一行即为 header 的描述。
其中每一列的格式为 `<tag_name/edge_name>.<prop_name>:<prop_type>`：

- `<tag_name/edge_name>` 表示 TAG 或者 EDGE 的名字，
- `<prop_name>` 表示属性名字，
- `<prop_type>` 表示属性类型，即上述中的 `bool`、`int`、`float`、`double`、`string` 和 `timestamp`。如果不设置默认为 `string`。

在上述的 `<prop_type>` 中有如下几个关键词含有特殊语义：

- `:VID` 表示顶点的 VID
- `:SRC_VID` 表示边的起点的 VID
- `:DST_VID` 表示边的终点的 VID
- `:RANK` 表示边的 rank 值
- `:IGNORE` 表示忽略这一列
- `:LABEL` 表示插入/删除 `+/-` 的标记列

数据文件含有 header，那么配置文件中的 tags/edge 下的 props 会被自动忽略，按照数据文件中的 header 属性设置插入数据。

#### 顶点

example 中 course 顶点的示例：

```csv
:LABEL,:VID,course.name,building.name:string,:IGNORE,course.credits:int
+,"hash(""Math"")",Math,No5,1,3
+,"uuid(""English"")",English,"No11 B\",2,6
```

因为 VERTEX 可以含有多个不同的 TAG，所以在指定对应的 column 的 header 时要加上 TAG 的 name。

在 `:VID` 这列除了常见的整数值，还可以使用 `hash` 和 `uuid` 两个 built-in 函数来自动产生顶点的 VID。需要注意的是在 CSV 文件中字符串的转义处理，如示例中的 `"hash(""Math"")"` 存储的是 `hash("Math")` 字符串。

上述中除了 `:LABEL` 这列（可选）之外，其他的列都可按任意顺序排列，对于已经存在的 CSV 文件而言，通过设置 header 便能灵活的选取自己需要的列来导入。

#### 边

example 中 follow 边的示例：

```csv
:DST_VID,follow.likeness:double,:SRC_VID,:RANK
201,92.5,200,0
200,85.6,201,1
```

上例中分别在第 0 列和第 2 列上指定为 follow 边的起点和终点的 VID 数据，最后一列为边的 rank 值。这些列的排列顺序同上述顶点一样，亦可自由排列，不过这仅限于带 header 的数据文件。

## Label

为了表示数据文件中的一行数据是进行插入还是删除操作，引入两个 label（+/-）符号。

- `+` 表示插入
- `-` 表示删除

这两个符号单独一列存储。

具体对应的示例如上述中带 header 的 vertex所示。

## 使用

### 源码

Nebula Importer 使用 **>=1.13** 版本的 golang 编译，所以首选确保你在系统中安装了上述的 golang 运行环境。安装和配置教程参考[文档](docs/golang-install.md)。

使用 `git` 克隆该仓库到本地，进入 `nebula-importer/cmd` 目录，直接执行即可。

``` shell
$ git clone https://github.com/vesoft-inc/nebula-importer.git
$ cd nebula-importer/cmd
$ go run importer.go --config /path/to/yaml/config/file
```

其中 `--config` 用来传入 YAML 配置文件的路径。

### Docker

使用 docker 可以不必在本地安装 golang 环境。直接拉取 Nebula Importer 的[镜像](https://hub.docker.com/r/vesoft/nebula-importer "nebula importer docker image")来导入，唯一要做的就是将本地配置文件和 CSV 数据文件挂载到容器中，如下所示：

```shell
$ docker run --rm -ti \
    --network=host \
    -v {your-config-file}:{your-config-file} \
    -v {your-csv-data-dir}:{your-csv-data-dir} \
    vesoft/nebula-importer
    --config {your-config-file}
```

- `{your-config-file}`：替换成你的本地配置文件的绝对路径，
- `{your-csv-data-dir}`：替换成你的本地 CSV 数据文件的绝对路径。

> 注意：`{your-csv-data-dir}` 需要同你的 YAML 配置中的 `files.path` 保持一致。

## TODO

- [X] Summary statistics of response
- [X] Write error log and data
- [X] Configure file
- [X] Concurrent request to Graph server
- [ ] Create space and tag/edge automatically
- [ ] Configure retry option for Nebula client
- [X] Support edge rank
- [X] Support label for add/delete(+/-) in first column
- [X] Support column header in first line
- [X] Support vid partition
- [X] Support multi-tags insertion in vertex
- [X] Provide docker image and usage
- [X] Make header adapt to props order defined in schema of configure file
- [X] Handle string column in nice way
- [ ] Update concurrency and batch size online
- [ ] Count duplicate vids
- [X] Support VID generation automatically
- [X] Output logs to file
