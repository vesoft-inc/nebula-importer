# Nebula-importer

## Introduction

[Nebula Graph](https://github.com/vesoft-inc/nebula-docker-compose) csv importer with `go`. This tool reads local csv files and writes into Nebula storage.

You can use this tool by source code or by docker.

> You should start a Nebula server by [`docker-compose`](https://github.com/vesoft-inc/nebula-docker-compose "nebula-docker-compose") or [rpm installation](https://github.com/vesoft-inc/nebula/tree/master/docs/manual-EN/3.build-develop-and-administration/3.deploy-and-administrations/deployment).  And also make sure the corrsponding space, tags and edge types have been created in Nebula.

## Prepare configure file

Nebula-importer will read a `YAML` configuration file to get information about connection to graph server, tag/edge schema, etc.

Here's an [example](example/example.yaml) of configuration file.

See description below

```yaml
version: v1rc1
description: example
settings:
  concurrency: 4 # number of graph clients
  connection:
    user: user
    password: password
    address: 127.0.0.1:3699
files:
  - path: ./edge.csv
    batchSize: 100
    type: csv
    csv:
      withHeader: false
      withLabel: false
    schema:
      space: test
      type: edge
      edge:
        name: edge_name
        withRanking: true
        props:
          - name: prop_name
            type: string
    error:
      failDataPath: ./err/edge.csv
      logPath: ./err/edge.log
  - path: ./vertex.csv
    batchSize: 100
    type: csv
    csv:
      withHeader: false
      withLabel: false
    schema:
      space: test
      type: vertex
      vertex:
        tags:
          - name: tag1
            props:
              - name: prop1
                type: int
              - name: prop2
                type: timestamp
          - name: tag2
            props:
              - name: prop3
                type: double
              - name: prop4
                type: string
    error:
      failDataPath: ./err/vertex.csv
      logPath: ./err/vertex.log
```

As for this example, nebula-importer will import two **csv** data files `edge.csv` and `vertex.csv` in turn.

### Configuration Properties

| options                                      | description                                                               | default        |
| :--                                          | :--                                                                       | :--            |
| version                                      | Configuration file version                                                | v1rc1          |
| description                                  | Description of this configure file                                        | ""             |
| settings                                     | Graph client settings                                                     | -              |
| settings.concurrency                         | Number of graph clients                                                   | 4              |
| settings.connection                          | Connection options of graph client                                        | -              |
| settings.connection.user                     | Username                                                                  | user           |
| settings.connection.password                 | Password                                                                  | password       |
| settings.connection.address                  | Address of graph client                                                   | 127.0.0.1:3699 |
| files                                        | File list to be imported                                                  | -              |
| files[0].path                                | File path                                                                 | ""             |
| files[0].type                                | File type                                                                 | csv            |
| files[0].csv                                 | CSV file options                                                          | -              |
| files[0].csv.withHeader                      | Whether csv file has header                                               | false          |
| files[0].csv.withLabel                       | Whether csv file has `+/-` label to represent **delete/insert** operation | false          |
| files[0].schema                              | Schema definition for this file data                                      | -              |
| files[0].schema.space                        | Space name created in nebula                                              | ""             |
| files[0].schema.type                         | Schema type: vertex or edge                                               | vertex         |
| files[0].schema.edge                         | Edge options                                                              | -              |
| files[0].schema.edge.name                    | Edge name in above space                                                  | ""             |
| files[0].schema.edge.withRanking             | Whether this edge has ranking                                             | false          |
| files[0].schema.edge.props                   | Properties of the edge                                                    | -              |
| files[0].schema.edge.props[0].name           | Property name                                                             | ""             |
| files[0].schema.edge.props[0].type           | Property type                                                             | ""             |
| files[0].schema.vertex                       | Vertex options                                                            | -              |
| files[0].schema.vertex.tags                  | Vertex tags options                                                       | -              |
| files[0].schema.vertex.tags[0].name          | Vertex tag name                                                           | ""             |
| files[0].schema.vertex.tags[0].props         | Vertex tag's properties                                                   | -              |
| files[0].schema.vertex.tags[0].props[0].name | Vertex tag's property name                                                | ""             |
| files[0].schema.vertex.tags[0].props[0].type | Vertex tag's property type                                                | ""             |
| files[0].error                               | File error configuration                                                  | -              |
| files[0].error.failDataPath                  | Failed data file path                                                     | ""             |
| files[0].error.logPath                       | Error log file path                                                       | ""             |

## CSV Data Example

There will be two csv data formats supported in the future. But now please use the first format which has no header line in your csv data file.

### Without Header Line

#### Vertex

In vertex csv data file, first column can be label(+/-) or vid. Vertex VID column always is the first column or following the label. And property values are behind VID.

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

This feature has not been supported now. Please remove the header from your csv data file at present.

#### Edge

```csv
1 _src,_dst,_ranking,prop1,prop2
2 ...
```

`_src` and `_dst` represent edge source and destination vertex id. `_ranking` column is value of edge ranking.

#### Vertex

```csv
1 _vid,tag1.prop1,tag2.prop2,tag1.prop3,tag2.prop4
2 ...
```

`_vid` column represent the global unique vertex id.

## Usage

### From Sources

This tool depends on golang 1.13, so make sure you have install `go` first.

Use `git` to clone this project to your local directory and execute the `cmd/importer.go` with `config` parameter.

``` shell
$ git clone https://github.com/yixinglu/nebula-importer.git
$ cd nebula-importer/cmd
$ go run importer.go --config /path/to/yaml/config/file
```

### Docker

With docker, we can easily to import our local data to nebula without `golang` runtime environment.

```shell
$ docker run --rm -ti \
    --network=host \
    -v {your-config-file}:/root/{your-config-file} \
    -v {your-csv-data-dir}:/root/{your-csv-data-dir} \
    xl4times/nebula-importer
    --config /root/{your-config-file}
```

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
