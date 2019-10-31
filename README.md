# Nebula-importer

## Introduction

`Nebula Graph Importer` with Go. This tool reads local csv files and writes into Nebula.

You can use this tool by source code, or by docker.

> You should start a Nebula server or [by `docker-compose`](https://github.com/vesoft-inc/nebula-docker-compose "nebula-docker-compose").  And also make sure the corrsponding tag/vertex or edge type have been created in Nebula.

## Prepare configure file

Nebula-importer will read the configuration file to get information about connection to graph server, schemas tag/vertex, etc.

Here's an [example](example/example.yaml) of configuration file.

See description below

```yaml
version: v1rc1
description: example
settings:
  retry: 3
  concurrency: 4 # Graph client pool size
  connection:
    user: user
    password: password
    address: 127.0.0.1:3699
files:
  - path: ~/example/edge.csv
    batchSize: 5
    type: csv
    csv:
      withHeader: false
    schema:
      space: example
      type: edge
      edge:
        name: edge
        props:
          - name: prop
            type: string
    error:
      failDataPath: ~/example/err/edge.csv
      logPath: ~/example/err/edge.log
  - path: ~/example/vertex.csv
    batchSize: 5
    type: csv
    csv:
      withHeader: true
    schema:
      space: example
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
      failDataPath: ~/example/err/vertex.csv
      logPath: ~/example/err/vertex.log
```

As for this example, nebula-importer will import two data source files inherit.csv(edges) and job.csv(vertexs) in turn.

### Configuration Properties

| options                       | description                          | default        |
| :--                           | :--                                  | :--            |
| version                       | Configure file version               | v1rc1          |
| description                   | Description of this configure file   | ""             |
| settings                      | Graph client settings                |                |
| settings.concurrency          | Number of clients                    | 4              |
| settings.retry                | Retry times when insert fails        | 3              |
| settings.connection           | Connection options of graph client   |                |
| settings.connection.user      | Username                             | user           |
| settings.connection.password  | Password                             | password       |
| settings.connection.address   | Address of graph client              | 127.0.0.1:3699 |
| files                         | File list to be imported             |                |
| files[0].path                 | File path                            | ""             |
| files[0].type                 | File type                            | csv            |
| files[0].schema               | Schema definition for this file data |                |
| files[0].schema.space         | Space name created in nebula         | ""             |
| files[0].schema.name          | Tag/Edge name in above space         | ""             |
| files[0].schema.type          | Schema type: vertex or edge          | vertex         |
| files[0].schema.props         | Properties of the schema             |                |
| files[0].schema.props[0].name | Property name                        | ""             |
| files[0].schema.props[0].type | Property type                        | ""             |

## Usage

### From Sources

This tool depends on golang 1.13, so make sure you have install `go` first.

Use `git` to clone this project to your local directory and execute the `main.go` with `config` parameter.

``` shell
$ git clone https://github.com/yixinglu/nebula-importer.git
$ cd nebula-importer
$ go run main.go --config /path/to/yaml/config/file
```

### Docker

With docker, we can easily to import our local data to nebula without `golang` runtime environment.

```shell
$ docker run --rm -ti \
    --network=host \
    -v {your-config-file}:/home/nebula/{your-config-file} \
    -v {your-csv-data-dir}:/home/nebula/{your-csv-data-dir} \
    xl4times/nebula-importer
    --config /home/nebula/{your-config-file}
```

## TODO

- [ ] Summary statistics of response
- [X] Write error log and data
- [X] Configure file
- [X] Concurrent request to Graph server
- [ ] Create space and tag/edge automatically
- [ ] Configure retry option for Nebula client
- [ ] Support edge rank
- [ ] Support label for add/delete(+/-) in first column
- [ ] Support column header in first line
- [ ] Support vid partition
- [ ] Support multi-tags insertion in vertex
- [X] Provide docker image and usage
- [ ] Make header adapt to props order defined in schema of configure file
- [ ] Handle string column in nice way
