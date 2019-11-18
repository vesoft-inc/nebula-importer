# Nebula Importer Configuration Description

| options                                        | description                                                               | default        |
| :--                                            | :--                                                                       | :--            |
| version                                        | Configuration file version                                                | v1rc1          |
| description                                    | Description of this configure file                                        | ""             |
| clientSettings                                 | Graph client settings                                                     | -              |
| clientSettings.concurrency                     | Number of graph clients                                                   | 4              |
| clientSettings.channelBufferSize               | Buffer size of client channels                                            | 128            |
| clientSettings.space                           | Space name of all data to be inserted                                     | ""             |
| clientSettings.connection                      | Connection options of graph client                                        | -              |
| clientSettings.connection.user                 | Username                                                                  | user           |
| clientSettings.connection.password             | Password                                                                  | password       |
| clientSettings.connection.address              | Address of graph client                                                   | 127.0.0.1:3699 |
| logPath                                        | Path of log file                                                          | ""             |
| files                                          | File list to be imported                                                  | -              |
| files[0].path                                  | File path                                                                 | ""             |
| files[0].type                                  | File type                                                                 | csv            |
| files[0].csv                                   | CSV file options                                                          | -              |
| files[0].csv.withHeader                        | Whether csv file has header                                               | false          |
| files[0].csv.withLabel                         | Whether csv file has `+/-` label to represent **delete/insert** operation | false          |
| files[0].schema                                | Schema definition for this file data                                      | -              |
| files[0].schema.type                           | Schema type: vertex or edge                                               | vertex         |
| files[0].schema.edge                           | Edge options                                                              | -              |
| files[0].schema.edge.name                      | Edge name in above space                                                  | ""             |
| files[0].schema.edge.withRanking               | Whether this edge has ranking                                             | false          |
| files[0].schema.edge.props                     | Properties of the edge                                                    | -              |
| files[0].schema.edge.props[0].name             | Property name                                                             | ""             |
| files[0].schema.edge.props[0].type             | Property type                                                             | ""             |
| files[0].schema.vertex                         | Vertex options                                                            | -              |
| files[0].schema.vertex.tags                    | Vertex tags options                                                       | -              |
| files[0].schema.vertex.tags[0].name            | Vertex tag name                                                           | ""             |
| files[0].schema.vertex.tags[0].props           | Vertex tag's properties                                                   | -              |
| files[0].schema.vertex.tags[0].props[0].name   | Vertex tag's property name                                                | ""             |
| files[0].schema.vertex.tags[0].props[0].type   | Vertex tag's property type                                                | ""             |
| files[0].failDataPath                          | Failed data file path                                                     | ""             |
