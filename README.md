<div align="center">
  <h1>Nebula Importer</h1>
  <div>
    <a href="https://github.com/vesoft-inc/nebula-importer/blob/master/README_zh-CN.md">中文</a>
  </div>
</div>

[![test](https://github.com/vesoft-inc/nebula-importer/workflows/test/badge.svg)](https://github.com/vesoft-inc/nebula-importer/actions?workflow=test)

## Introduction

Nebula Importer is a CSV import tool for [Nebula Graph](https://github.com/vesoft-inc/nebula) by [Docker compose](https://github.com/vesoft-inc/nebula-docker-compose). It can read and import data in local CSV files.

Before you try Nebula Importer, make sure you have:

* Deployed Nebula Graph
* Created schema for the space to import data

Currently, there are two ways to deploy Nebula:

1. [nebula-docker-compose](https://github.com/vesoft-inc/nebula-docker-compose "nebula-docker-compose")
2. [rpm Package](https://github.com/vesoft-inc/nebula/tree/master/docs/manual-EN/3.build-develop-and-administration/3.deploy-and-administrations/deployment)

> The easiest way to try Nebula Graph is using [`docker-compose`](https://github.com/vesoft-inc/nebula-docker-compose).

## Prepare Configuration File

Nebula-importer reads the CSV file to be imported and Nebula server data through the YAML configuration file.Here's an [example](example/example.yaml) of the configuration file and the CSV file. Detail descriptions for the configuration file see the following section.

```yaml
version: v1rc1
description: example
```

* `version` is a **required** parameter that indicates the configure file's version, the default version is `v2rc1`.
* `description` is an **optional** parameter that describes the configure file.
* `clientSettings` takes care all the Nebula related configurations.

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

* `clientSettings.concurrency` is an optional parameter that shows the concurrency of Nebula Graph Client, i.e. the connection number of Nebula Graph Server, the default value is 10.
* `clientSettings.channelBufferSize` is an optional parameter that shows the buffer size of cache queue for each Nebula Graph Client, the default value is 128.
* `clientSettings.space` is a **required** parameter that specifies which `space` the data will be importing into. Do not import data to multiple spaces at one time for the performance sake.
* `clientSettings.connection` is a **required** parameter that contains the `user`, `password` and `address` information of Nebula Graph Server.

### Files

The log and data file related configurations are:

* `logPath`: **Optional**. Specifies log directory when importing data.
* `files`: **Required**. An array type to configure different CSV files.

```yaml
logPath: ./err/test.log
files:
  - path: ./edge.csv
    failDataPath: ./err/edge.csv
    batchSize: 100
    type: csv
    csv:
      withHeader: false
      withLabel: false
```

### CSV Data Files

One CSV file can only store one type of vertex or edge. Vertices and edges of different schema should be stored in different files.

* `path`: **Required**. Specifies the path where the CSV data file is stored. If a relative path is used, the path and directory of the current configuration file are spliced.
* `failDataPath`: **Required**. Specifies the file to insert the failed data output so that the error data is appended later.
* `batchSize`: **Optional**. Specifies the batch size of the inserted data, the default value is 128.
* `type & csv`:  **Required**. Specifies the file type. Currently only CSV is supported. You can specify whether to include the header and the inserted and deleted labels in the CSV file.
  * `withHeader`: The default value is false, the format of the header is described below.
  * `withLabel`: The default value is false, the format of the label is described below.

* `schema`: **Required**. Describes the metadata information of the current data file. The schema.type has only two values: vertex and edge.
  * When type is specified as vertex, details should be described in vertex field.
  * When type is specified as edge, details should be described in edge field.

`schema.vertex`

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

`schema.vertex` is a **required** parameter that describes the schema information such as tags of the inserted vertex. Since sometimes one vertex contains several tags, different tags should be given in the `schema.vertex.tags` array.

Each tag contains the following two properties:

* `name`: The tag's name.
* `prop`: The tag's properties. Each property contains the following two fields:
  * `name`: The property name, the same with the tag property in Nebula Graph
  * `type`: The property type, currently support `bool`, `int`, `float`, `double`, `timestamp` and `string`.

> Note: The order of properties in the above props must be the same as that of the corresponding data in the CSV data file.

`schema.edge`

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

`schema.edge` is a **required** parameter that describes the schema information of the inserted edge. Each edge contains the following three properties:

* `name`: The edge's name.
* `ranking`: Specifies the `ranking` value of the given edge, used to tell different edges share the same edge type and vertices.
* `prop`: Same as the above tag. Please be noted the property order here must be the same with that of the corresponding data in the CSV data file.

Details of all the configurations please refer to [Configuration Reference](docs/configuration-reference.md).

## About the CSV Header

Usually you can add some descriptions in the first row of the CSV file to specify each column's type.

### Data Without Header

If the `csv.withHeader` is set to `false`, the CSV file only contains the data (no descriptions of the first row). Example of vertices and edges are as follow:

#### Vertex Example

Take vertex course in example:

```csv
101,Math,3,No5
102,English,6,No11
```

The first column is the vertex ID, the following three columns are the properties, corresponding to the course.name, course.credits and building.name in the configuration file. (See  `vertex.tags.props`).

#### Edge Example

Take edge choose in example:

```csv
200,101,5
200,102,3
```

### Data With Header

If the `csv.withHeader` is set to `false`, the CSV file only contains the data (no descriptions of the first row). Example of vertices and edges are as follow:

The first two columns indicate source vertex and dest vertex ID, the third is the property, corresponding to choose.likeness in the configuration file. (If ranking is included, put the values in the third column. The properties should follow behind ranking column.)

## CSV Data Example

There will be two csv data formats supported in the future. But now please use the first format which has no header line in your csv data file.

### With Header Line

If the `csv.withHeader` is set to `true`, the first row of the CSV file is header.
The format of each column is `<tag_name/edge_name>.<prop_name>:<prop_type>`:

* `<tag_name/edge_name>` is the name of the vertex or edge.
* `<prop_name>` is the property name.
* `<prop_type>` is the property type. It can be `bool`, `int`, `float`, `double`, `string` and `timestamp`, the default type is `string`.

In the above `<prop_type>` field, the following keywords contain special semantics:

* `:VID` is the vertex ID.
* `:SRC_VID` is the source vertex VID.
* `:DST_VID` is the dest vertex VID.
* `:RANK` is the rank of the edge.
* `:IGNORE` indicates this column will be ignored.
* `:LABEL` indicates the columns that insert/delete `+/-`.

> If the csv file contains header, the importer parses the schema of each row according to the header and ignores the `props` in YAML.

#### Example of Vertex CSV File With Header

Take vertex course as example:

```csv
:LABEL,:VID,course.name,building.name:string,:IGNORE,course.credits:int
+,"hash(""Math"")",Math,No5,1,3
+,"uuid(""English"")",English,"No11 B\",2,6
```

##### LABEL (Optional)

```csv
:LABEL,
+,
-,
```

Indicates the column is inserting (+) or deleting (-) operation.

##### :VID (Required）

```csv
:VID
123,
"hash(""Math"")",
"uuid(""English"")"
```

In the `:VID` column, in addition to the common integer values ​​(such as 123), you can also use the two built-in functions `hash` and `uuid` to automatically calculate the VID of the generated vertex (for example, hash("Math")).

> Note that the double quotes (") are escaped in the CSV file. For example, `hash("Math")` should be written as `"hash(""Math"")"`.

##### Other Properties

```csv
course.name,:IGNORE,course.credits:int
Math,1,3
English,2,6
```

You can specify `:IGNORE` to ignore the second column when importing data. In addition, all columns except the `:LABEL` column can be in any order. This way, for a large csv file, you can flexibly select the columns you need by setting the header.

> Because a VERTEX can contain multiple TAGs, the TAG name should be added to the header of the specified column (for example, it must be `course.credits`, rather than the abbreviated `credits`).

#### Example of Edge CSV File With Header

Take edge follow as example:

```csv
:DST_VID,follow.likeness:double,:SRC_VID,:RANK
201,92.5,200,0
200,85.6,201,1
```

In the preceding example, the source vertex of the edge is `:SRC_VID` (in column 4), the dest vertex of the edge is `:DST_VID` (in column 1), and the property on the edge is `follow.likeness:double`(in column 2), the ranking field of the edge is `:RANK` (in column 5, the default value is 0 if you do not specify).

##### Label（Optional）

* `+` means inserting
* `-` means deleting

The same with vertex, you
You can specify label in edge CSV file header the same way with vertex.

## Use This Importer Tool by Source Code or Docker

After completing the configuration of the YAML file and the preparation of the (to be imported) csv data file, you can use this tool to batch write to Nebula.

### From Source code

Nebula Importer is compiled with golang higher than **>=1.13**, so make sure that golang is installed on your system. The installation and configuration tutorial is referenced in [documentation](docs/golang-install.md).

Use `git` to clone the repository to local, go to the `nebula-importer/cmd` directory and execute the importer.

``` bash
$ git clone https://github.com/vesoft-inc/nebula-importer.git
$ cd nebula-importer/cmd
$ go run importer.go --config /path/to/yaml/config/file
```

`--config` is used to pass in the path to the YAML configuration file.

### From Docker

With Docker you don't have to install golang locally. Pull Nebula Importer's [Docker Image](https://hub.docker.com/r/vesoft/nebula-importer) to import. The only thing to do is to mount the local configuration file and the CSV data file into the container as follows:

```bash
$ docker run --rm -ti \
    --network=host \
    -v {your-config-file}:{your-config-file} \
    -v {your-csv-data-dir}:{your-csv-data-dir} \
    vesoft/nebula-importer
    --config {your-config-file}
```

* `{your-config-file}`: Replace with the absolute path of the local YAML configuration file
* `{your-csv-data-dir}`: Replace with the absolute path of the local CSV data file.

> Note: It is recommended to use relative paths in files.path. But if you use a local absolute path in `files.path`, you need to carefully check the path mapped to Docker with this path.

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
