# 无表头配置说明

对于无表头（header）的CSV文件，需要在配置文件里设置`withHeader`为`false`，表示CSV文件中只含有数据（不含第一行表头），同时可能还需要设置数据类型、对应的列等。

## 示例文件

无表头的CSV文件示例如下：

- 点示例

  `student_without_header.csv`的示例数据：

  ```csv
  student100,Monica,16,female
  student101,Mike,18,male
  student102,Jane,17,female
  ```

  第一列为点ID，后面三列为属性`name`、`age`和`gender`。

- 边示例

  `follow_without_header.csv`的示例数据：

  ```csv
  student100,student101,0,92.5
  student101,student100,1,85.6
  student101,student102,2,93.2
  student100,student102,1,96.2
  ```

  前两列的数据分别为起始点ID和目的点ID，第三列为rank，第四列为属性`degree`。

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
  - path: ./student_without_header.csv

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
      withHeader: false

      # 是否有LABEL。
      withLabel: false

      # 指定csv文件的分隔符。只支持一个字符的字符串分隔符。
      delimiter: ","

    schema:
      # Schema的类型，可选值为vertex和edge。
      type: vertex

      vertex:
        
        # 点ID设置。
        vid:
           # 点ID对应CSV文件中列的序号。CSV文件中列的序号从0开始。
           index: 0

           # 点ID的数据类型，可选值为int和string，分别对应Nebula Graph中的INT64和FIXED_STRING。
           type: string

        # 标签设置。   
        tags:
            # 标签名称。
          - name: student
           
            # 标签内的属性设置。
            props:
                # 属性名称。
              - name: name
                
                # 属性数据类型。
                type: string

                # 属性对应CSV文件中列的序号。
                index: 1

              - name: age
                type: int
                index: 2
              - name: gender
                type: string
                index: 3

    # 本示例第二个数据文件为边的数据。
  - path: ./follow_without_header.csv
    failDataPath: ./err/followerr.csv
    batchSize: 10
    limit: 10
    inOrder: true
    type: csv
    csv:
      withHeader: false
      withLabel: false
    schema:
      # Schema的类型为edge。
      type: edge
      edge:
        # 边类型名称。
        name: follow

        # 是否包含rank。
        withRanking: true

        # 起始点ID设置。
        srcVID:
           # 数据类型。
           type: string

           # 起始点ID对应CSV文件中列的序号。
           index: 0

        # 目的点ID设置。
        dstVID:
           type: string
           index: 1

        # rank设置。
        rank:
           # rank值对应CSV文件中列的序号。如果没有设置index，请务必在第三列设置rank的值。之后的列依次设置各属性。
           index: 2
        
        # 边类型内的属性设置。
        props:
             # 属性名称。
           - name: degree
             
             # 属性数据类型。
             type: double

             # 属性对应CSV文件中列的序号。
             index: 3
```

- CSV文件中列的序号从0开始，即第一列的序号为0，第二列的序号为1。

- 点ID的数据类型需要和`clientSettings.postStart.commands`中的创建图空间语句的数据类型一致。

- 如果没有设置index字段指定列的序号，CSV文件必须遵守如下规则：

  - 在点数据文件中，第一列必须为点ID，后面的列为属性，且需要和配置文件内的顺序一一对应。

  - 在边数据文件中，第一列必须为起始点ID，第二列必须为目的点ID，如果`withRanking`为`true`，第三列必须为rank值，后面的列为属性，且需要和配置文件内的顺序一一对应。