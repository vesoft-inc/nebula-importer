# Nebula Importer Configuration Description

| options                                       | description                                                               | default        |
| :--                                           | :--                                                                       | :--            |
| version                                       | Configuration file version                                                | v1             |
| description                                   | Description of this configure file                                        | ""             |
| removeTempFiles                               | Whether to remove generated temporary data and log files                  | false          |
| clientSettings                                | Graph client settings                                                     | -              |
| clientSettings.retry                          | Number of graph clients retry to execute failed nGQL                      | 1              |
| clientSettings.concurrency                    | Number of graph clients                                                   | 4              |
| clientSettings.channelBufferSize              | Buffer size of client channels                                            | 128            |
| clientSettings.space                          | Space name of all data to be inserted                                     | ""             |
| clientSettings.connection                     | Connection options of graph client                                        | -              |
| clientSettings.connection.user                | Username                                                                  | user           |
| clientSettings.connection.password            | Password                                                                  | password       |
| clientSettings.connection.address             | Address of graph client                                                   | 127.0.0.1:9669 |
| clientSettings.postStart.commands             | Post scripts after connecting nebula                                      | ""             |
| clientSettings.postStart.afterPeriod          | The period time between running post scripts and inserting data           | 0s             |
| clientSettings.preStop.commands               | Prescripts before disconnecting nebula                                    | ""             |
| logPath                                       | Path of log file                                                          | ""             |
| files                                         | File list to be imported                                                  | -              |
| files[0].path                                 | File path                                                                 | ""             |
| files[0].failDataPath                         | Failed data file path                                                     | ""             |
| files[0].batchSize                            | Size of each batch for inserting stmt construction                        | 128            |
| files[0].limit                                | Limit rows to be read                                                     | NULL           |
| files[0].inOrder                              | Whether to insert rows in order                                           | false          |
| files[0].type                                 | File type                                                                 | csv            |
| files[0].csv                                  | CSV file options                                                          | -              |
| files[0].csv.withHeader                       | Whether csv file has header                                               | false          |
| files[0].csv.withLabel                        | Whether csv file has `+/-` label to represent **delete/insert** operation | false          |
| files[0].csv.delimiter                        | The delimiter of csv file to separate different columns                   | ","            |
| files[0].schema                               | Schema definition for this file data                                      | -              |
| files[0].schema.type                          | Schema type: vertex or edge                                               | vertex         |
| files[0].schema.edge                          | Edge options                                                              | -              |
| files[0].schema.edge.srcVID.index             | Column index of source vertex id of edge                                  | 0              |
| files[0].schema.edge.srcVID.function          | The generation function of edge source vertex id                          | ""             |
| files[0].schema.edge.dstVID.index             | Column index of destination vertex id of edge                             | 1              |
| files[0].schema.edge.dstVID.function          | The generation function of edge destination vertex id                     | ""             |
| files[0].schema.edge.rank.index               | Column index of the edge rank                                             | 2              |
| files[0].schema.edge.name                     | Edge name in above space                                                  | ""             |
| files[0].schema.edge.props                    | Properties of the edge                                                    | -              |
| files[0].schema.edge.props[0].name            | Property name                                                             | ""             |
| files[0].schema.edge.props[0].type            | Property type                                                             | ""             |
| files[0].schema.edge.props[0].index           | Property index                                                            |                |
| files[0].schema.vertex                        | Vertex options                                                            | -              |
| files[0].schema.vertex.vid.index              | Column index of vertex vid                                                | 0              |
| files[0].schema.vertex.vid.function           | The generation function of vertex vid                                     | ""             |
| files[0].schema.vertex.tags                   | Vertex tags options                                                       | -              |
| files[0].schema.vertex.tags[0].name           | Vertex tag name                                                           | ""             |
| files[0].schema.vertex.tags[0].props          | Vertex tag's properties                                                   | -              |
| files[0].schema.vertex.tags[0].props[0].name  | Vertex tag's property name                                                | ""             |
| files[0].schema.vertex.tags[0].props[0].type  | Vertex tag's property type                                                | ""             |
| files[0].schema.vertex.tags[0].props[0].index | Vertex tag's property index                                               |                |
