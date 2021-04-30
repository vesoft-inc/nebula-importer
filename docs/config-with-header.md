# 有表头配置说明

对于有表头（header）的CSV文件，需要在配置文件里设置`withHeader`为`true`，表示CSV文件中第一行为表头，表头内容具有特殊含义。

>**警告**：如果CSV文件中含有header，Importer就会按照header来解析每行数据的Schema，并忽略yaml文件中的点或边设置。

## 示例文件

有表头的CSV文件示例如下：

- 点示例

  `student_with_header.csv`的示例数据：

  ```csv
  :VID(string),student.name:string,student.age:int,student.gender:string
  student100,Monica,16,female
  student101,Mike,18,male
  student102,Jane,17,female
  ```

  第一列为点ID，后面三列为属性`name`、`age`和`gender`。

- 边示例

  `follow_with_header.csv`的示例数据：

  ```csv
  :SRC_VID(string),:DST_VID(string),:RANK,follow.degree:double
  student100,student101,0,92.5
  student101,student100,1,85.6
  student101,student102,2,93.2
  student100,student102,1,96.2
  ```

  前两列的数据分别为起始点ID和目的点ID，第三列为rank，第四列为属性`degree`。

## 表头格式说明

表头通过一些关键词定义起始点、目的点、rank以及一些特殊功能，说明如下：

- `:VID`（必填）：点ID。需要用`:VID(type)`形式设置数据类型，例如`:VID(string)`或`:VID(int)`。

- `:SRC_VID`（必填）：边的起始点ID。需要用`:SRC_VID(type)`形式设置数据类型。

- `:DST_VID`（必填）：边的目的点ID。需要用`:DST_VID(type)`形式设置数据类型。

- `:RANK`（可选）：边的rank值。

- `:IGNORE`（可选）：插入数据时忽略这一列。

- `:LABEL`（可选）：表示对该行进行插入（+）或删除（-）操作。必须为第一列。例如：

  ```csv
  :LABEL,
  +,
  -,
  ```

>**说明**：除了`:LABEL`列之外的所有列都可以按任何顺序排序，因此针对较大的CSV文件，您可以灵活地设置header来选择需要的列。

对于标签或边类型的属性，格式为`<tag_name/edge_name>.<prop_name>:<prop_type>`，说明如下：

- `<tag_name/edge_name>`：标签或者边类型的名称。

- `<prop_name>`：属性名称。

- `<prop_type>`：属性类型。支持`bool`、`int`、`float`、`double`、`timestamp`和`string`，默认为`string`。

例如`student.name:string`、`follow.degree:double`。



## 配置示例

```yaml
# 连接的Nebula Graph版本，连接2.x时设置为v2。
version: v2

description: example

# 是否删除临时生成的日志和错误数据文件。
removeTempFiles: false

clientSettings:

  # nGQL语句执行失败的重试次数。
  retry: 3

  # Nebula Graph客户端并发数。
  concurrency: 10 

  # 每个Nebula Graph客户端的缓存队列大小。
  channelBufferSize: 128

  # 指定数据要导入的Nebula Graph图空间。
  space: student

  # 连接信息。
  connection:
    user: root
    password: nebula
    address: 192.168.*.*:9669


  postStart:
    # 配置连接Nebula Graph服务器之后，在插入数据之前执行的一些操作。
    commands: |
      DROP SPACE IF EXISTS student;
      CREATE SPACE IF NOT EXISTS student(partition_num=5, replica_factor=1, vid_type=FIXED_STRING(20));
      USE student;
      CREATE TAG student(name string, age int,gender string);
      CREATE EDGE follow(degree int);

    # 执行上述命令后到执行插入数据命令之间的间隔。
    afterPeriod: 15s
  
  preStop:
    # 配置断开Nebula Graph服务器连接之前执行的一些操作。
    commands: |

# 错误等日志信息输出的文件路径。    
logPath: ./err/test.log

# CSV文件相关设置。
files:
  
    # 数据文件的存放路径，如果使用相对路径，则会将路径和当前配置文件的目录拼接。本示例第一个数据文件为点的数据。
  - path: ./student_with_header.csv

    # 插入失败的数据文件存放路径，以便后面补写数据。
    failDataPath: ./err/studenterr.csv

    # 单批次插入数据的语句数量。
    batchSize: 10

    # 读取数据的行数限制。
    limit: 10

    # 是否按顺序在文件中插入数据行。如果为false，可以避免数据倾斜导致的导入速率降低。
    inOrder: true

    # 文件类型，当前仅支持csv。
    type: csv

    csv:
      # 是否有表头。
      withHeader: true

      # 是否有LABEL。
      withLabel: false

      # 指定csv文件的分隔符。只支持一个字符的字符串分隔符。
      delimiter: ","

    schema:
      # Schema的类型，可选值为vertex和edge。
      type: vertex

    # 本示例第二个数据文件为边的数据。
  - path: ./follow_with_header.csv
    failDataPath: ./err/followerr.csv
    batchSize: 10
    limit: 10
    inOrder: true
    type: csv
    csv:
      withHeader: true
      withLabel: false
    schema:
      # Schema的类型为edge。
      type: edge
      edge:
        # 边类型名称。
        name: follow

        # 是否包含rank。
        withRanking: true
```

点ID的数据类型需要和`clientSettings.postStart.commands`中的创建图空间语句的数据类型一致。