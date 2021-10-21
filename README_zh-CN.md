<div align="center">
  <h1>Nebula Importer</h1>
  <div>
    <a href="https://github.com/vesoft-inc/nebula-importer/blob/master/README.md">EN</a>
  </div>
</div>

[![test](https://github.com/vesoft-inc/nebula-importer/workflows/test/badge.svg)](https://github.com/vesoft-inc/nebula-importer/actions?workflow=test)

<!--
## 介绍

Nebula Importer 是一款 [Nebula Graph](https://github.com/vesoft-inc/nebula) 的 CSV 文件导入工具，其读取本地的 CSV 文件，然后写入到 Nebula Graph 图数据库中。

在使用 Nebula Importer 之前，首先请确保：

- Nebula Graph 服务已部署。
- 元数据信息 `space`、`tag` 和 `edge` 已创建好。

目前有三种部署方式：

1. [nebula-docker-compose](https://github.com/vesoft-inc/nebula-docker-compose "nebula-docker-compose")
2. [rpm 包安装](https://docs.nebula-graph.com.cn/manual-CN/3.build-develop-and-administration/2.install/1.install-with-rpm-deb/)
3. [源码编译安装](https://docs.nebula-graph.com.cn/manual-CN/3.build-develop-and-administration/1.build/1.build-source-code/)

> 如果想在本地快速试用 Nebula Graph，推荐使用 `docker-compose` 部署。

## 如何使用

在完成 YAML 配置文件和（待导入）CSV 数据文件准备后，就可以使用本工具向 Nebula Graph 批量写入数据。

### 源码编译方式

Nebula Importer 使用 **1.13** 或更新版本的 Go 语言编译，所以首先确保在系统中安装了上述的 golang 运行环境。安装和配置教程参考[文档](docs/golang-install.md)。

使用 `git` 克隆该仓库到本地，进入 `nebula-importer/` 目录，运行 `make build`。

``` bash
$ git clone https://github.com/vesoft-inc/nebula-importer.git
$ cd nebula-importer
$ make build
$ ./nebula-importer --config /path/to/yaml/config/file
```

其中 `--config` 用来传入 YAML 配置文件的路径。

### Docker 方式

使用 Docker 可以不必在本地安装 Go 语言环境。直接拉取 Nebula Importer 的[镜像](https://hub.docker.com/r/vesoft/nebula-importer)来导入。唯一要做的就是将本地配置文件和 CSV 数据文件挂载到容器中，如下所示：

```bash
$ docker run --rm -ti \
    --network=host \
    -v {your-config-file}:{your-config-file} \
    -v {your-csv-data-dir}:{your-csv-data-dir} \
    vesoft/nebula-importer
    --config {your-config-file}
```

- `{your-config-file}`：替换成本地 YAML 配置文件的绝对路径
- `{your-csv-data-dir}`：替换成本地 CSV 数据文件的绝对路径

> 注意：通常建议在 `files.path` 中使用相对路径。但如果在 `files.path` 中使用本地绝对路径，则需要小心检查这个路径映射到 Docker 中的对应路径。

## 配置文件

Nebula Importer 通过 YAML 配置文件来描述要导入的文件信息、Nebula Graph 的 server 信息等。[这里](examples/)有一个配置文件的参考样例和对应的数据文件格式。接下来逐一解释各个选项的含义：

```yaml
version: v2
description: example
removeTempFiles: false
```

- `version`：**必填**。表示配置文件的版本，默认值为 `v2`。
- `description`：**可选**。对当前配置文件的描述信息。
- `removeTempFiles`：**可选**。是否删除生成的临时日志和错误数据文件，默认值为：`false`。
- `clientSettings`：跟 Nebula Graph 服务端相关的配置均在该字段下配置。

```yaml
clientSettings:
  retry: 3
  concurrency: 10
  channelBufferSize: 128
  space: test
  connection:
    user: user
    password: password
    address: 192.168.8.1:9669,192.168.8.2:9669
  postStart:
    commands: |
      UPDATE CONFIGS storage:wal_ttl=3600;
      UPDATE CONFIGS storage:rocksdb_column_family_options = { disable_auto_compactions = true };
    afterPeriod: 8s
  preStop:
    commands: |
      UPDATE CONFIGS storage:wal_ttl=86400;
      UPDATE CONFIGS storage:rocksdb_column_family_options = { disable_auto_compactions = false };
```

- `clientSettings.retry`：**可选**。表示 Nebula Graph Client 的重试失败的 nGQL 请求次数，默认为 1。
- `clientSettings.concurrency`：**可选**。表示 Nebula Graph Client 的并发度，即同 Nebula Graph Server 的连接数，默认为 10。
- `clientSettings.channelBufferSize`：**可选**。表示每个 Nebula Graph Client 对应的缓存队列 (channel) 的 buffer 大小，默认为 128。
- `clientSettings.space`：**必填**。指定所有的数据文件将要导入到哪个 `space`。请不要同时向多个 space 批量导入数据，这样反而性能更低。
- `clientSettings.connection`：**必填**。配置 Nebula Graph Server 的 `user`、`password` 和 `address` 信息。
- `clientSettings.postStart`：**可选**。配置连接 Nebula Graph Server 之后，在插入数据之前执行的一些操作。
  - `clientSettings.postStart.commands`：定义连接 Nebula Graph Server 之后的一些命令。
  - `clientSettings.postStart.afterPeriod`：定义执行上述命令之后到真正插入数据之前的间隔。
- `clientSettings.preStop`：**可选**。配置断开 Nebula Graph Server 连接之前执行的一些操作。
  - `clientSettings.preStop.commands`：定义断开连接 Nebula Graph Server 之前的一些命令脚本。

### 文件

跟日志和数据文件相关的配置跟以下两个选项有关：

- `logPath`：**可选**。指定导入过程中的错误等日志信息输出的文件路径，默认输出到 `/tmp/nebula-importer-{timestamp}.log` 中。
- `files`：**必填**。数组类型，用来配置不同的数据文件。您也可以从 HTTP 链接导入数据，在文件路径中输入链接即可。

```yaml
logPath: ./err/test.log
files:
  - path: ./student.csv
    failDataPath: ./err/student.csv
    batchSize: 128
    limit: 10
    inOrder: false
    type: csv
    csv:
      withHeader: false
      withLabel: false
      delimiter: ","
```

#### 数据文件

一个数据文件中只能存放一种点或者边，不同 schema 的点或者边数据需要放置在不同的文件中。

- `path`：**必填**。指定数据文件的存放路径，如果使用相对路径，则会拼接当前配置文件的目录和 `path`。
- `failDataPath`：**必填**。指定插入失败的数据输出的文件，以便后面补写出错数据。
- `batchSize`：**可选**。批量插入数据的条数，默认 128。
- `limit`：**可选**。限制读取文件的行数。
- `inOrder`：**可选**。是否按序插入文件中的每一行。如果不指定，可以避免数据倾斜导致的导入速率的下降。
- `type` & `csv`：**必填**。指定文件的类型，目前只支持 CSV 文件导入。在 CSV 文件中可以指定是否含有文件头和插入、删除的标记。
  - `withHeader`：默认是 `false`，文件头的格式在后面描述。
  - `withLabel`：默认是 `false`，label 的格式也在后面描述。
  - `delimiter`：**可选**。指定 CSV 文件的分隔符，默认是 `","`。目前只有单字符的分隔符被支持。

#### `schema`

**必填**。描述当前数据文件的元数据信息。`schema.type` 只有两种取值：`vertex` 和 `edge`。

- 当指定 `type: vertex` 时，需要在 `vertex` 字段中继续描述。
- 当指定 `type: edge` 时，需要在 `edge` 字段中继续描述。

##### `schema.vertex`

**必填**。描述插入点的 schema 信息，比如 tags。

```yaml
schema:
  type: vertex
  vertex:
    vid:
      index: 1
      function: hash
    tags:
      - name: student
        props:
          - name: age
            type: int
            index: 2
          - name: name
            type: string
            index: 1
          - name: gender
            type: string
```

##### `schema.vertex.vid`

**可选**。描述点 VID 所在的列和使用的函数。

- `index`：**可选**。在 CSV 文件中的列标，从 0 开始计数。默认值 0。
- `function`：**可选**。用来生成 VID 时的函数，有 `hash` 和 `uuid` 两种函数可选。

##### `schema.vertex.tags`

**可选**。由于一个 VERTEX 可以含有多个 TAG，所以不同的 TAG 在 `schema.vertex.tags` 数组中给出。

对于每一个 TAG，有以下两个属性:

- `name`：TAG 的名称。
- `props`：TAG 的属性字段数组，每个属性字段又由如下两个字段构成：
  - `name`：**必填**。属性名称，同 Nebula Graph 中创建的 TAG 的属性名称一致。
  - `type`：**必填**。属性类型，目前支持 `bool`、`int`、`float`、`double`、`timestamp`、`string`、`geography`、`geography(point)`、`geography(linestring)`和`geography(polygon)` 几种类型。
  - `index`：**可选**。在 CSV 文件中的列标。

> **注意**：上述 `props` 中的属性描述**顺序**必须同数据文件中的对应数据排列顺序一致。

##### `schema.edge`

**必填**。描述插入边的 schema 信息。

```yaml
schema:
  type: edge
  edge:
    name: choose
    srcVID:
      index: 0
      function: hash
    dstVID:
      index: 1
      function: uuid
    rank:
      index: 2
    props:
      - name: grade
        type: int
        index: 3
```

含有如下字段：

- `name`：**必填**。边的名称，同 Nebula Graph 中创建的 edge 名称一致。
- `srcVID`：**可选**。边的起点信息，含有的 `index` 和 `function` 意义同上述 `vertex.vid`。
- `dstVID`：**可选**。边的终点信息，含有的 `index` 和 `function` 意义同上述 `vertex.vid`。
- `rank`：**可选**。边的 rank 信息，含有的 `index` 表示该值所在的列。
- `props`：**必填**。描述同上述点，同样需要注意跟数据文件中列的排列顺序一致。

所有配置的选项解释见[表格](docs/configuration-reference.md)。

## 关于 CSV Header

通常还可以在 CSV 文件的第一行添加一些描述信息，以指定每列的类型。

### 没有 header 的数据格式

如果在上述配置中的 `csv.withHeader` 配置为 `false`，那么 CSV 文件中只含有数据（不含有第一行描述信息）。对于点和边的数据示例如下：

#### 点示例

example 中 course 点的样例数据：

```csv
101,Math,3,No5
102,English,6,No11
```

第一列为点的 `VID`。后面三列为属性值，分别按序对应配置文件中的 course.name、course.credits 和 building.name（见 `vertex.tags.props`）。

#### 边示例

example 中 choose 类型的边的样例数据：

```csv
200,101,5
200,102,3
```

前两列的数据分别为起点 VID 和终点 VID，第三列对应 choose.likeness 属性（如果边中含有 rank 字段，请在第三列放置 rank 的值。之后的列依次放置各属性）。

### 含有 header 的数据格式

如果配置文件中 `csv.withHeader` 设置为 `true`，那么对应的数据文件中的第一行即为 header 的描述。

每一列的格式为 `<tag_name/edge_name>.<prop_name>:<prop_type>`：

- `<tag_name/edge_name>` 表示 TAG 或者 EDGE 的名称。
- `<prop_name>` 表示属性名称。
- `<prop_type>` 表示属性类型。可以是 `bool`、`int`、`float`、`double`、`string`、`timestamp`、`geography`、`geography(point)`、`geography(linestring)`和`geography(polygon)`，不设置默认为 `string`。

在上述的 `<prop_type>` 字段中有如下几个关键词含有特殊语义：

- `:VID` 表示点的 VID
- `:SRC_VID` 表示边的起点的 VID
- `:DST_VID` 表示边的终点的 VID
- `:RANK` 表示边的 rank 值
- `:IGNORE` 表示忽略这一列
- `:LABEL` 表示插入/删除 `+/-` 的标记列

> **注意**：如果 csv 文件中含有 header 描述信息，那么工具就按照会 header 来解析每行数据的 schema，并忽略 YAML 中的 `props`。

#### 含有 header 的点 csv 文件示例

example 中 course 点的示例：

```csv
:LABEL,:VID,course.name,building.name:string,:IGNORE,course.credits:int
+,"hash(""Math"")",Math,No5,1,3
+,"uuid(""English"")",English,"No11 B\",2,6
```

##### LABEL (可选）

```csv
:LABEL,
+,
-,
```

表示该行为插入(+)或者删除(-)操作。

##### :VID (必选）

```csv
:VID
123,
"hash(""Math"")",
"uuid(""English"")"
```

在 `:VID` 这列除了常见的整数值（例如 123），还可以使用 `hash` 和 `uuid` 两个内置函数来自动计算生成点的 VID（例如 hash("Math")）。

> 需要注意的是在 CSV 文件中对双引号(")的转义处理。如 `hash("Math")` 要写成 `"hash(""Math"")"`。

##### 其他属性

```csv
course.name,:IGNORE,course.credits:int
Math,1,3
English,2,6
```

可以指明 `:IGNORE` 表示忽略第二列不需要导入。此外，除了 `:LABEL` 这列之外，其他的列都可按任意顺序排列。这样对于一个比较大的 CSV 文件，可以通过设置 header 来灵活地选取自己需要的列。

> **注意**：因为 VERTEX 可以含有多个不同的 TAG，所以在指定列的 header 时要加上 TAG 名称（例如必须是 `course.credits`，不能简写为 `credits`）。

#### 含有 header 的边 csv 文件示例

example 中 follow 边的示例：

```csv
:DST_VID,follow.likeness:double,:SRC_VID,:RANK
201,92.5,200,0
200,85.6,201,1
```

可以看到，例子中边的起点为 `:SRC_VID`（在第 4 列），边的终点为 `:DST_VID`（在第 1 列），边上的属性为 `follow.likeness:double`（在第 2 列），边的 rank 字段对应 `:RANK`（在第 5 列，如果不指定导入 `:RANK` 则系统默认为 0）。

#### Label（可选）

- `+` 表示插入
- `-` 表示删除

边 CSV 文件 header 中也可以指定 label，和点原理相同。

-->
