<div align="center">
  <h1>Nebula Importer</h1>
  <div>
    <!--
    <a href="https://github.com/vesoft-inc/nebula-importer/blob/master/README_zh-CN.md">中文</a>
    -->
  </div>
</div>

[![test](https://github.com/vesoft-inc/nebula-importer/workflows/test/badge.svg)](https://github.com/vesoft-inc/nebula-importer/actions?workflow=test)

## Introduction

Nebula Importer is a CSV importing tool for [Nebula Graph](https://github.com/vesoft-inc/nebula). It reads data in the local CSV files and imports data into Nebula Graph.

Before you start Nebula Importer, make sure that:

* Nebula Graph is deployed.
* A schema, composed of space, tags, and edge types, is created.

Currently, there are three methods to deploy Nebula Graph:

1. [nebula-docker-compose](https://github.com/vesoft-inc/nebula-docker-compose "nebula-docker-compose")
2. [rpm package](https://docs.nebula-graph.io/3.1.0/4.deployment-and-installation/2.compile-and-install-nebula-graph/2.install-nebula-graph-by-rpm-or-deb/)
3. [from source](https://docs.nebula-graph.io/3.1.0/4.deployment-and-installation/2.compile-and-install-nebula-graph/1.install-nebula-graph-by-compiling-the-source-code/)

> The quickest way to deploy Nebula Graph is using [Docker Compose](https://github.com/vesoft-inc/nebula-docker-compose).

## **CAUTION**: Choose the correct branch

The rpc protocols (i.e., thrift) in Nebula Graph 1.x, v2, v3 are incompatible.
Nebula Importer master and v3 branch can only connect to Nebula Graph 3.x.

> Do not mismatch.

## How to use

After configuring the YAML file and preparing the CSV files to be imported, you can use this tool to batch write data to Nebula Graph.

### From the source code

Nebula Importer is compiled with Go **1.13** or later, so make sure that Go is installed on your system. See the Go [installation document](docs/golang-install-en.md) for the installation and configuration tutorial.

1. Clone the repository

  * For Nebula Graph 3.x, clone the master branch.

  ``` bash
  $ git clone https://github.com/vesoft-inc/nebula-importer.git
  ```

2. Go to the `nebula-importer` directory.

```
$ cd nebula-importer
```

3. Build the source code.

```
$ make build
```

4. Start the service

```
$ ./nebula-importer --config /path/to/yaml/config/file
```

The `--config` option in the preceding command is used to pass the path of the YAML configuration file.

### From Docker

If you are using Docker, you don't have to install Go locally. Pull the [Docker image](https://hub.docker.com/r/vesoft/nebula-importer) for Nebula Importer. Mount the local configuration file and the CSV data files into the container and you are done.

```bash
$ docker run --rm -ti \
    --network=host \
    -v {your-config-file}:{your-config-file} \
    -v {your-csv-data-dir}:{your-csv-data-dir} \
    vesoft/nebula-importer:{image_version}
    --config {your-config-file}
```

- `{your-config-file}`: Replace with the absolute path of the local YAML configuration file.
- `{your-csv-data-dir}`: Replace with the absolute path of the local CSV data file.
- `{image_version}`: Replace with the image version you need(e.g. `v1`, `v2`, `v3`)
> **NOTE**: We recommend that you use the relative paths in the `files.path` file. If you use the local absolute path, check how the path is mapped to Docker carefully.

## Prepare the configuration file

Nebula Importer uses the YAML configuration file to store information for the CSV files and Nebula Graph server. Here's an [example for v2](examples/v2/example.yaml) and an [example for v1](examples/v1/example.yaml) for the configuration file and the CSV file. You can find the explanation for each option in the following:

```yaml
version: v2
description: example
removeTempFiles: false
```

* `version`: **Required**. Indicates the configuration file version, the default value is `v2`. Note that `v2` config can be used with both 2.x and 3.x Nebula service.
* `description`: **Optional**. Describes the configuration file.
* `removeTempFiles`: **Optional**. Whether to delete the temporarily generated log and error data files. The default value is `false`.
* `clientSettings`: Stores all the configurations related to the Nebula Graph service.

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

* `clientSettings.retry`: **Optional**. Shows the failed retrying times to execute nGQL queries in Nebula Graph client.
* `clientSettings.concurrency`: **Optional**. Shows the concurrency of Nebula Graph Client, i.e. the connection number between the Nebula Graph Client and the Nebula Graph Server. The default value is 10.
* `clientSettings.channelBufferSize`: **Optional**. Shows the buffer size of the cache queue for each Nebula Graph Client, the default value is 128.
* `clientSettings.space`: **Required**. Specifies which `space` the data is imported into. Do not import data to multiple spaces at the same time because it causes a performance problem.
* `clientSettings.connection`: **Required**. Configures the `user`, `password`, and `address` information for Nebula Graph Server.
* `clientSettings.postStart`: **Optional**. Stores the operations that are performed after the Nebula Graph Server is connected and before any data is inserted.
  * `clientSettings.postStart.commands`: Defines some commands that will run when Nebula Graph Server is connected.
  * `clientSettings.postStart.afterPeriod`: Defines the interval between running the preceding commands and inserting data to Nebula Graph Server.
* `clientSettings.preStop`: **Optional**. Configures the operations before disconnecting Nebula Graph Server.
  * `clientSettings.preStop.commands`: Defines some command scripts before disconnecting Nebula Graph Server.

### Files

The following three configurations are related to the log and data files:

* `workingDir`: **Optional**. If you have multiple directories containing data with the same file structure, you can use this parameter to switch between them. For example, the value of `path` and `failDataPath` of the configuration below will be automatically changed to `./data/student.csv` and `./data/err/student.csv`. If you change workingDir to `./data1`, the path will be changed accordingly. The param can be either absolute or relative.
* `logPath`: **Optional**. Specifies the log path when importing data. The default path is `/tmp/nebula-importer-{timestamp}.log`.
* `files`: **Required**. It is an array type to configure different data files. You can also import data from a HTTP link by inputting the link in the file path.

```yaml
workingDir: ./data/
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

#### CSV data files

One CSV file can only store one type of vertex or edge. Vertices and edges of the different schema must be stored in different files.

* `path`: **Required**. Specifies the path where the data files are stored. If a relative path is used, the `path` and current configuration file directory are spliced. Wildcard filename is also supported, for example: `./follower-*.csv`, please make sure that all matching files with the same schema.
* `failDataPath`: **Required**. Specifies the path for data that failed in inserting so that the failed data are reinserted.
* `batchSize`: **Optional**. Specifies the batch size of the inserted data. The default value is 128.
* `limit`: **Optional**. Limits the max data reading rows.
* `inOrder`: **Optional**. Whether to insert the data rows in the file in order. If you do not specify it, you avoid the decrease in importing rate caused by the data skew.

* `type & csv`: **Required**. Specifies the file type. Currently, only CSV is supported. Specify whether to include the header and the inserted and deleted labels in the CSV file.
  * `withHeader`: The default value is false. The format of the header is described in the following section.
  * `withLabel`: The default value is false. The format of the label is described in the following section.
  * `delimiter`: **Optional**. Specify the delimiter for the CSV files. The default value is `","`. And only a 1-character string delimiter is supported.

#### `schema`

**Required**. Describes the metadata information for the current data file. The `schema.type` has only two values: vertex and edge.

* When type is set to vertex, details must be described in the vertex field.
* When type is set to edge, details must be described in edge field.

##### `schema.vertex`

**Required**. Describes the schema information for vertices. For example, tags.

```yaml
schema:
  type: vertex
  vertex:
    vid:
      index: 1
      function: hash
      prefix: abc
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

**Optional**. Describes the vertex ID column and the function used for the vertex ID.

* `index`: **Optional**. The column number in the CSV file. Started with 0. The default value is 0.
* `function`: **Optional**. Functions to generate the VIDs. Currently, we only support function `hash` and `uuid`.
* `type`: **Optional**. The type for VIDs. The default value is `string`.
* `prefix`: **Optional**. Add prefix to the original vid. When `function` is specified also, `prefix` is applied to the original vid before `function`.

##### `schema.vertex.tags`

**Optional**. Because a vertex can have several tags, different tags are described in the `schema.vertex.tags` parameter.

Each tag contains the following two properties:

* `name`: The tag name.
* `prop`: A property of the tag. Each property contains the following two fields:
  * `name`: **Required**. The property name, must be the same with the tag property in Nebula Graph.
  * `type`: **Optional**. The property type, currently  `bool`, `int`, `float`, `double`, `string`, `time`, `timestamp`, `date`, `datetime`, `geography`, `geography(point)`, `geography(linestring)` and `geography(polygon)` are supported.
  * `index`: **Optional**. The column number in the CSV file.

> **NOTE**: The properties in the preceding `prop` parameter must be sorted in the **same** way as in the CSV data file.

##### `schema.edge`

**Required**. Describes the schema information for edges.

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

The edge parameter contains the following fields:

* `name`: **Required**. The name of the edge type.
* `srcVID`: **Optional**. The source vertex information for the edge. The `index` and `function` included here are the same as that of in the `vertex.vid` parameter.
* `dstVID`: **Optional**. The destination vertex information for the edge. The `index` and `function` included here are the same as that of in the `vertex.vid` parameter.
* `rank`: **Optional**. Specifies the `rank` value for the edge. The `index` indicates the column number in the CSV file.
* `props`: **Required**. The same as the `props` in the vertex. The properties in the `prop` parameter must be sorted in the **same** way as in the CSV data file.

See the [Configuration Reference](docs/configuration-reference.md) for details on the configurations.

## About the CSV header

Usually, you can add some descriptions in the first row of the CSV file to specify the type for each column.

### Data without header

If the `csv.withHeader` is set to `false`, the CSV file only contains the data (no descriptions in the first row). Example for vertices and edges are as follows:

#### Vertex example

Take tag `course` for example:

```csv
101,Math,3,No5
102,English,6,No11
```

The first column is the vertex ID, the following three columns are the properties, corresponding to the course.name, course.credits and building.name in the configuration file. (See  `vertex.tags.props`).

#### Edge example

Take edge type `choose` for example:

```csv
200,101,5
200,102,3
```

The first two columns are the source VID and destination VID. The third column corresponds to the choose.likeness property. If an edge contains the rank value, put it in the third column. Then put the edge properties in order.

### Data with header

If the `csv.withHeader` is set to `true`, the first row of the CSV file is the header information.

The format for each column is `<tag_name/edge_name>.<prop_name>:<prop_type>`:

* `<tag_name/edge_name>` is the name for the vertex or edge.
* `<prop_name>` is the property name.
* `<prop_type>` is the property type. It can be `bool`, `int`, `float`, `double`, `string`, `time`, `timestamp`, `date`, `datetime`, `geography`, `geography(point)`, `geography(linestring)` and `geography(polygon)`. The default type is `string`.

In the above `<prop_type>` field, the following keywords contain special semantics:

* `:VID` is the vertex ID.
* `:SRC_VID` is the source vertex VID.
* `:DST_VID` is the destination vertex VID.
* `:RANK` is the rank of the edge.
* `:IGNORE` indicates the column is ignored.
* `:LABEL` indicates the column is marked as inserted/deleted `+/-`.

> **NOTE**: If the CSV file contains the header, the importer parses the schema of each row according to the header and ignores the `props` in YAML.

#### Example of vertex CSV file with header

Take vertex course as example:

```csv
:LABEL,:VID,course.name,building.name:string,:IGNORE,course.credits:int
+,"hash(""Math"")",Math,No5,1,3
+,"uuid(""English"")",English,"No11 B\",2,6
```

##### LABEL (optional)

```csv
:LABEL,
+,
-,
```

Indicates the column is the insertion (+) or deletion (-) operation.

##### :VID (required）

```csv
:VID
123,
"hash(""Math"")",
"uuid(""English"")"
```

In the `:VID` column, in addition to the common integer values (such as 123), you can also use the two built-in functions `hash` and `uuid` to automatically generate the VID for the vertices (for example, hash("Math")).

> **NOTE**: The double quotes (") are escaped in the CSV file. For example, `hash("Math")` must be written as `"hash(""Math"")"`.

##### Other Properties

```csv
course.name,:IGNORE,course.credits:int
Math,1,3
English,2,6
```

`:IGNORE` is to specify the column that you want to ignore when importing data. All columns except the `:LABEL` column can be sorted in any order. Thus, for a large CSV file, you can select the columns you need flexibly by setting the header.

> Because a VERTEX can contain multiple TAGs, when specifying the header, you must add the tag name. For example, it must be `course.credits`, rather than the abbreviated `credits`).

#### Example of edge CSV file with header

Take edge `follow` for example:

```csv
:DST_VID,follow.likeness:double,:SRC_VID,:RANK
201,92.5,200,0
200,85.6,201,1
```

In the preceding example, the source vertex of the edge is `:SRC_VID` (in column 4), the destination vertex of the edge is `:DST_VID` (in column 1), and the property on the edge is `follow.likeness:double`(in column 2), the ranking field of the edge is `:RANK` (in column 5. The default value is 0 if you do not specify).

#### Label（optional）

* `+` means inserting
* `-` means deleting

Similar to vertex, you can specify label for header in the edge CSV file .
