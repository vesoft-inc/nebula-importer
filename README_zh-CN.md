
<p align="center">
  <img src="https://github.com/vesoft-inc/nebula/raw/master/docs/logo.png"/>
  <br> <a href="README.md">English</a> | 中文
  <br>A distributed, scalable, lightning-fast graph database<br>
</p>
<div align="center">
  <h1>Nebula Importer</h1>
</div>
<p align="center">
  <a href="http://githubbadges.com/star.svg?user=vesoft-inc&repo=nebula&style=default">
    <img src="http://githubbadges.com/star.svg?user=vesoft-inc&repo=nebula&style=default" alt="nebula star"/>
  </a>
  <a href="http://githubbadges.com/fork.svg?user=vesoft-inc&repo=nebula-graph&style=default">
    <img src="http://githubbadges.com/fork.svg?user=vesoft-inc&repo=nebula&style=default" alt="nebula fork"/>
  </a>
  <br>
</p>

Nebula Importer（简称Importer）是一款[Nebula Graph](https://github.com/vesoft-inc/nebula)的CSV文件导入工具。Importer可以读取本地的CSV文件，然后导入数据至Nebula Graph图数据库中。

## 适用场景

Importer适用于将本地CSV文件的内容导入至Nebula Graph中。

## 优势

- 轻量快捷：不需要复杂环境即可使用，快速导入数据。

- 灵活筛选：通过配置文件可以实现对CSV文件数据的灵活筛选。

## 前提条件

在使用Nebula Importer之前，请确保：

- 已部署Nebula Graph服务。目前有三种部署方式：
  
  - [Docker Compose部署](https://docs.nebula-graph.com.cn/2.0/2.quick-start/2.deploy-nebula-graph-with-docker-compose/)（快速部署）
  
  - [RPM/DEB包安装](https://docs.nebula-graph.com.cn/2.0/4.deployment-and-installation/2.compile-and-install-nebula-graph/2.install-nebula-graph-by-rpm-or-deb/)
  
  - [源码编译安装](https://docs.nebula-graph.com.cn/2.0/4.deployment-and-installation/2.compile-and-install-nebula-graph/1.install-nebula-graph-by-compiling-the-source-code/)

- Nebula Graph中已创建Schema，包括图空间、标签和边类型，或者通过参数`clientSettings.postStart.commands`设置。

- 运行Importer的机器已部署Golang环境。详情请参见[Golang 环境搭建](docs/golang-install.md)。

## 操作步骤

配置yaml文件并准备好待导入的CSV文件，即可使用本工具向Nebula Graph批量写入数据。

### 源码编译运行

1. 克隆仓库。

   ```bash
   $ git clone --branch <branch> https://github.com/vesoft-inc/nebula-importer.git
   ```

   >**说明**：请使用正确的分支。 
   >
   >Nebula Graph 1.x和2.x的rpc协议不同，因此：
   >
   >- Nebula Importer v1分支只能连接Nebula Graph 1.x。
   >
   >- Nebula Importer master分支和v2分支可以连接Nebula Graph 2.x。

2. 进入目录`nebula-importer`。

   ```bash
   $ cd nebula-importer
   ```

3. 编译源码。

   ```bash
   $ make build
   ```

4. 启动服务。

   ```bash
   $ ./nebula-importer --config <yaml_config_file_path>
   ```

   >**说明**：yaml配置文件说明请参见[配置文件](#配置文件说明)。

#### 无网络编译方式

如果您的服务器不能联网，建议您在能联网的机器上将源码和各种以来打包上传到对应的服务器上编译即可，操作步骤如下：

1. 克隆仓库。

   ```bash
   $ git clone --branch <branch> https://github.com/vesoft-inc/nebula-importer.git
   ```

2. 使用如下的命令下载并打包依赖的源码。

   ```bash
   $ cd nebula-importer
   $ go mod vendor
   $ cd .. && tar -zcvf nebula-importer.tar.gz nebula-importer
   ```

3. 将压缩包上传到不能联网的服务器上。

4. 解压并编译。

   ```bash
   $ tar -zxvf nebula-importer.tar.gz 
   $ cd nebula-importer
   $ go build -mod vendor cmd/
   ```

### Docker方式运行

使用Docker可以不必在本地安装Go语言环境，只需要拉取Nebula Importer的[镜像](https://hub.docker.com/r/vesoft/nebula-importer)，并将本地配置文件和CSV数据文件挂载到容器中。命令如下：

```bash
$ docker run --rm -ti \
    --network=host \
    -v <config_file>:<config_file> \
    -v <csv_data_dir>:<csv_data_dir> \
    vesoft/nebula-importer:<version>
    --config <config_file>
```

- `<config_file>`：本地yaml配置文件的绝对路径。
- `<csv_data_dir>`：本地CSV数据文件的绝对路径。
- `<version>`：Nebula Graph 2.x请填写`v2`。

> **说明**：建议您使用相对路径。如果使用本地绝对路径，请检查路径映射到Docker中的路径。



## 配置文件说明

Nebula Importer通过yaml配置文件来描述待导入文件信息、Nebula Graph服务器信息等。您可以参考示例配置文件：[2.0配置文件](examples/v2/example.yaml)/[1.0配置文件](examples/v1/example.yaml)。下文将分类介绍配置文件内的字段。

### 基本配置

示例配置如下：

```yaml
version: v2
description: example
removeTempFiles: false
```

|参数|默认值|是否必须|说明|
|:---|:---|:---|:---|
|`version`|v2|是|目标Nebula Graph的版本。|
|`description`|example|否|配置文件的描述。|
|`removeTempFiles`|false|否|是否删除临时生成的日志和错误数据文件。|

### 客户端配置

客户端配置存储客户端连接Nebula Graph相关的配置。

示例配置如下：

```yaml
clientSettings:
  retry: 3
  concurrency: 10
  channelBufferSize: 128
  space: test
  connection:
    user: user
    password: password
    address: 192.168.*.*:9669,192.168.*.*:9669
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

|参数|默认值|是否必须|说明|
|:---|:---|:---|:---|
|`clientSettings.retry`|3|否|nGQL语句执行失败的重试次数。|
|`clientSettings.concurrency`|10|否|Nebula Graph客户端并发数。|
|`clientSettings.channelBufferSize`|128|否|每个Nebula Graph客户端的缓存队列大小。|
|`clientSettings.space`|-|是|指定数据要导入的Nebula Graph图空间。不要同时导入多个空间，以免影响性能。|
|`clientSettings.connection.user`|-|是|Nebula Graph的用户名。|
|`clientSettings.connection.password`|-|是|Nebula Graph用户名对应的密码。|
|`clientSettings.connection.address`|-|是|所有Graph服务的地址和端口。|
|`clientSettings.postStart.commands`|-|否|配置连接Nebula Graph服务器之后，在插入数据之前执行的一些操作。|
|`clientSettings.postStart.afterPeriod`|-|否|执行上述`commands`命令后到执行插入数据命令之间的间隔，例如`8s`。|
|`clientSettings.preStop.commands`|-|否|配置断开Nebula Graph服务器连接之前执行的一些操作。|

### 文件配置

文件配置存储数据文件和日志的相关配置，以及Schema的具体信息。

#### 文件和日志配置

示例配置如下：

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

|参数|默认值|是否必须|说明|
|:---|:---|:---|:---|
|`logPath`|`/tmp/nebula-importer-<timestamp>.log`|否|导入过程中的错误等日志信息输出的文件路径。|
|`files.path`|`./student.csv`|是|数据文件的存放路径，如果使用相对路径，则会将路径和当前配置文件的目录拼接。|
|`files.failDataPath`|`./err/student.csv`|是|插入失败的数据文件存放路径，以便后面补写数据。|
|`files.batchSize`|128|否|单批次插入数据的语句数量。|
|`files.limit`|-|否|读取数据的行数限制。|
|`files.inOrder`|-|否|是否按顺序在文件中插入数据行。如果为`false`，可以避免数据倾斜导致的导入速率降低。|
|`files.type`|-|是|文件类型。|
|`files.csv.withHeader`|`false`|是|是否有表头。详情请参见[关于CSV文件表头](#关于csv文件表头header)。|
|`files.csv.withLabel`|`false`|是|是否有LABEL。详情请参见[含有header的数据格式](#含有header的数据格式)|
|`files.csv.delimiter`|`","`|是|指定csv文件的分隔符。只支持一个字符的字符串分隔符。|

#### Schema配置

Schema配置描述当前数据文件的Meta信息，Schema的类型分为点和边两类，可以同时配置多个点或边。

- 点配置

示例配置如下：

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

|参数|默认值|是否必须|说明|
|:---|:---|:---|:---|
|`files.schema.type`|-|是|Schema的类型，可选值为`vertex`和`edge`。|
|`files.schema.vertex.vid.index`|0|否|点ID对应CSV文件中列的序号。|
|`files.schema.vertex.vid.function`|-|否|通过函数生成点ID。支持的函数为`hash`和`uuid`。|
|`files.schema.vertex.tags.name`|-|是|标签名称|
|`files.schema.vertex.tags.props.name`|-|是|标签属性名称，必须和Nebula Graph中的标签属性一致。|
|`files.schema.vertex.tags.props.type`|-|否|属性类型，支持`bool`、`int`、`float`、`double`、`timestamp`和`string`。|
|`files.schema.vertex.tags.props.index`|-|否|属性对应CSV文件中列的序号。|

>**说明**：
>
>- 如果没有设置`index`字段，请确保`props`字段内的属性填写顺序和CSV文件内属性列的顺序一致。
>- CSV文件中列的序号从0开始，即第一列的序号为0，第二列的序号为1。

- 边配置

示例配置如下：

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

|参数|默认值|是否必须|说明|
|:---|:---|:---|:---|
|`files.schema.type`|-|是|Schema的类型，可选值为`vertex`和`edge`。|
|`files.schema.edge.name`|-|是|边类型名称。|
|`files.schema.edge.srcVID.index`|-|否|边的起始点ID对应CSV文件中列的序号。|
|`files.schema.edge.srcVID.function`|-|否|通过函数生成起始点ID。支持的函数为`hash`和`uuid`。|
|`files.schema.edge.dstVID.index`|-|否|边的目的点ID对应CSV文件中列的序号。|
|`files.schema.edge.dstVID.function`|-|否|通过函数生成目的点ID。支持的函数为`hash`和`uuid`。|
|`files.schema.edge.rank.index`|-|否|边的rank值对应CSV文件中列的序号。|
|`files.schema.edge.props.name`|-|是|边类型属性名称，必须和Nebula Graph中的边类型属性一致。|
|`files.schema.edge.props.type`|-|否|属性类型，支持`bool`、`int`、`float`、`double`、`timestamp`和`string`。|
|`files.schema.edge.props.index`|-|否|属性对应CSV文件中列的序号。|

>**说明**：
>
>- 如果没有设置`index`字段，请确保`props`字段内的属性填写顺序和CSV文件内属性列的顺序一致。
>- CSV文件中列的序号从0开始，即第一列的序号为0，第二列的序号为1。

## 关于CSV文件表头（header）

通常可以将CSV文件的第一行作为表头，添加特定的描述信息以指定每列的类型。

### 没有header的数据格式

如果配置中的`withHeader`为`false`，表示CSV文件中只含有数据（不含第一行表头）。

没有header的CSV文件示例如下：

- 点示例

  `student.csv`的样例数据：

  ```csv
  x200,Monica,16,female
  y201,Mike,18,male
  z202,Jane,17,female
  ```

  第一列为点ID，后面三列为属性值，按顺序分别对应配置文件中的`student.name`、`student.age`和`student.gender`。

- 边示例

  `choose.csv`的样例数据：

  ```csv
  x200,x101,5
  x200,y102,3
  y201,y102,3
  z202,y102,3
  ```

  前两列的数据分别为起始点ID和目的点ID，第三列为属性值，对应`choose.grade`。

  >**说明**：如果没有设置`index`字段且需要使用rank，请在第三列设置rank的值。之后的列依次设置各属性。

### 含有header的数据格式

如果配置中的`withHeader`为`true`，表示CSV文件中第一行为表头，表头内容具有特殊含义。

每一列的格式为`<tag_name/edge_name>.<prop_name>:<prop_type>`：

- `<tag_name/edge_name>`：标签或者边类型的名称。
- `<prop_name>`：属性名称。
- `<prop_type>`：属性类型。支持`bool`、`int`、`float`、`double`、`timestamp`和`string`，默认为`string`。

除此之外，表头还有如下几个关键词有特殊语义：

- `:VID`（必填）：点ID。可以用`:VID(type)`形式设置点ID的数据类型，例如`:VID(string)`或`:VID(int)`。除了常见的整数值（例如123），还可以使用`hash`和`uuid`两个内置函数来自动生成点ID。例如：

  ```csv
  :VID(string)
  123,
  "hash(""Math"")",
  "uuid(""English"")"
  ```

  >**说明**：双引号（"）在CSV文件中会被转义，例如`hash("Math")`必须写作`"hash(""Math"")"`。

- `:SRC_VID`：边的起始点ID。
- `:DST_VID`：边的目的点ID。
- `:RANK`：边的rank值。
- `:IGNORE`：插入数据时忽略这一列。
- `:LABEL`（可选）：表示对该行进行插入（+）或删除（-）操作。例如：

  ```csv
  :LABEL,
  +,
  -,
  ```

>**警告**：如果CSV文件中含有header，Importer就会按照header来解析每行数据的Schema，并忽略yaml文件中的`props`设置。

含有header的CSV文件示例如下：

- 点示例

  ```csv
  :LABEL,:VID,student.name,student.age,student.gender
  +,x200,Monica,16,female,2
  +,y201,Mike,18,male,5
  +,z202,Jane,17,female,7
  ```

- 边示例

  ```csv
  :SRC_VID,:DST_VID,choose.grade:int
  x200,x101,5
  x200,y102,3
  y201,y102,3
  z202,y102,3
  ```

>**说明**：
>
>- 除了`:LABEL`列之外的所有列都可以按任何顺序排序，因此针对较大的CSV文件，您可以灵活地设置header来选择需要的列。
>- 因为一个点可以包含多个标签，所以在设置header时，必须添加标签名称。例如`student.name`不能简写为`name`。