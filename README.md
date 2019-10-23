# nebula-importer

Nebula Graph Importer with Go

## Usage

### Configure file format

[example configure file](example/example.yaml)

```yaml
version: 1beta
description: example
settings:
  retry: 5
  concurrency: 4 # Graph client pool size
  connection:
    user: user
    password: password
    address: 127.0.0.1:3699
files:
  - path: ~/example/inherit.csv
    type: csv
    schema:
      space: sp
      name: inherit
      type: edge
      props:
        - name: job_id
          type: string
        - name: start_time
          type: timestamp
    error:
      failDataPath: ~/example/inherit/err/inherit.csv
      logPath: ~/example/inherit/err/inherit.log

  - path: ~/example/job.csv
    type: csv
    schema:
      space: sp
      name: job
      type: vertex
      props:
        - name: job_id
          type: string
        - name: start_time
          type: timestamp
    error:
      failDataPath: ~/example/job/err/job.csv
      logPath: ~/example/job/err/job.log
```

| options                       | description                          | default        |
| :--                           | :--                                  | :--            |
| version                       | Configure file version               | 1beta          |
| description                   | Description of this configure file   | ""             |
| settings                      | Graph client settings                |                |
| settings.concurrency          | Number clients                       | 4              |
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
| files[0].schema.name          | Vertex/Edge name in above space      | ""             |
| files[0].schema.type          | Schema type: vertex or edge          | vertex         |
| files[0].schema.props         | Properties of the schema             |                |
| files[0].schema.props[0].name | Property name                        | ""             |
| files[0].schema.props[0].type | Property type                        | ""             |

### Usage

``` shell
go run main.go --config /path/to/yaml/config/file
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
