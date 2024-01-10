# NebulaGraph Importer Configuration Description

| options                                     | description                                                                                          | default          |
|:--------------------------------------------|:-----------------------------------------------------------------------------------------------------|:-----------------|
| client                                      | The NebulaGraph client configuration options.                                                        | -                |
| client.version                              | Specifies which version of NebulaGraph, currently only `v3` is supported.                            | -                |
| client.address                              | The address of graph in NebulaGraph.                                                                 | -                |
| client.user                                 | The user of NebulaGraph.                                                                             | root             |
| client.password                             | The password of NebulaGraph.                                                                         | nebula           |
| client.handshakeKey                         | The handshakeKey of NebulaGraph.                                                                     | -                |
| client.ssl                                  | SSL related configuration.                                                                           | nebula           |
| client.ssl.enable                           | Specifies whether to enable ssl authentication.                                                      | false            |
| client.ssl.certPath                         | Specifies the path of the certificate file.                                                          | -                |
| client.ssl.keyPath                          | Specifies the path of the private key file.                                                          | -                |
| client.ssl.caPath                           | Specifies the path of the certification authority file.                                              | -                |
| client.ssl.insecureSkipVerify               | Specifies whether a client verifies the server's certificate chain and host name.                    | false            |
| client.concurrencyPerAddress                | The number of client connections to each graph in NebulaGraph.                                       | 10               |
| client.reconnectInitialInterval             | The initialization interval for reconnecting NebulaGraph.                                            | 1s               |
| client.retry                                | The failed retrying times to execute nGQL queries in NebulaGraph client.                             | 3                |
| client.retryInitialInterval                 | The initialization interval retrying.                                                                | 1s               |
|                                             |                                                                                                      |                  |
| manager                                     | The global control configuration options related to NebulaGraph Importer.                            | -                |
| manager.spaceName                           | Specifies which space the data is imported into.                                                     | -                |
| manager.batch                               | Specifies the batch size for all sources of the inserted data.                                       | 128              |
| manager.readerConcurrency                   | Specifies the concurrency of reader to read from sources.                                            | 50               |
| manager.importerConcurrency                 | Specifies the concurrency of generating statement, call client to import.                            | 512              |
| manager.statsInterval                       | Specifies the interval at which statistics are printed.                                              | 10s              |
| manager.hooks.before                        | Configures the statements before the import begins.                                                  | -                |
| manager.hooks.before.[].statements          | Defines the list of statements.                                                                      | -                |
| manager.hooks.before.[].wait                | Defines the waiting time after executing the above statements.                                       | -                |
| manager.hooks.after                         | Configures the statements after the import is complete.                                              | -                |
| manager.hooks.after.[].statements           | Defines the list of statements.                                                                      | -                |
| manager.hooks.after.[].wait                 | Defines the waiting time after executing the above statements.                                       | -                |
|                                             |                                                                                                      |                  |
| log                                         | The log configuration options.                                                                       | -                |
| log.level                                   | Specifies the log level.                                                                             | "INFO"           |
| log.console                                 | Specifies whether to print logs to the console.                                                      | true             |
| log.files                                   | Specifies which files to print logs to.                                                              | -                |
|                                             |                                                                                                      |                  |
| sources                                     | The data sources to be imported                                                                      | -                |
| sources[].path                              | Local file path                                                                                      | -                |
| sources[].s3.endpoint                       | The endpoint of s3 service.                                                                          | -                |
| sources[].s3.region                         | The region of s3 service.                                                                            | -                |
| sources[].s3.bucket                         | The bucket of file in s3 service.                                                                    | -                |
| sources[].s3.key                            | The object key of file in s3 service.                                                                | -                |
| sources[].s3.accessKeyID                    | The `Access Key ID` of s3 service.                                                                   | -                |
| sources[].s3.accessKeySecret                | The `Access Key Secret` of s3 service.                                                               | -                |
| sources[].oss.endpoint                      | The endpoint of oss service.                                                                         | -                |
| sources[].oss.bucket                        | The bucket of file in oss service.                                                                   | -                |
| sources[].oss.key                           | The object key of file in oss service.                                                               | -                |
| sources[].oss.accessKeyID                   | The `Access Key ID` of oss service.                                                                  | -                |
| sources[].oss.accessKeySecret               | The `Access Key Secret` of oss service.                                                              | -                |
| sources[].ftp.host                          | The host of ftp service.                                                                             | -                |
| sources[].ftp.host                          | The port of ftp service.                                                                             | -                |
| sources[].ftp.user                          | The user of ftp service.                                                                             | -                |
| sources[].ftp.password                      | The password of ftp service.                                                                         | -                |
| sources[].ftp.path                          | The path of file in the ftp service.                                                                 | -                |
| sources[].sftp.host                         | The host of sftp service.                                                                            | -                |
| sources[].sftp.host                         | The port of sftp service.                                                                            | -                |
| sources[].sftp.user                         | The user of sftp service.                                                                            | -                |
| sources[].sftp.password                     | The password of sftp service.                                                                        | -                |
| sources[].sftp.keyFile                      | The ssh key file path of sftp service.                                                               | -                |
| sources[].sftp.keyData                      | The ssh key file content of sftp service.                                                            | -                |
| sources[].sftp.passphrase                   | The ssh key passphrase of sftp service.                                                              | -                |
| sources[].sftp.path                         | The path of file in the ftp service.                                                                 | -                |
| sources[].hdfs.address                      | The address of hdfs service.                                                                         | -                |
| sources[].hdfs.user                         | The user of hdfs service.                                                                            | -                |
| sources[].hdfs.servicePrincipalName         | The kerberos service principal name of hdfs service when enable kerberos.                            | -                |
| sources[].hdfs.krb5ConfigFile               | The kerberos config file of hdfs service when enable kerberos.                                       | "/etc/krb5.conf" |
| sources[].hdfs.ccacheFile                   | The ccache file of hdfs service when enable kerberos.                                                | -                |
| sources[].hdfs.keyTabFile                   | The keytab file of hdfs service when enable kerberos.                                                | -                |
| sources[].hdfs.password                     | The kerberos password of hdfs service when enable kerberos.                                          | -                |
| sources[].hdfs.dataTransferProtection       | The data transfer protection of hdfs service.                                                        | -                |
| sources[].hdfs.disablePAFXFAST              | Whether to prohibit the client to use PA_FX_FAST.                                                    | -                |
| sources[].hdfs.path                         | The path of file in the sftp service.                                                                | -                |
| sources[].gcs.endpoint                      | The endpoint of GCS service.                                                                         | -                |
| sources[].gcs.bucket                        | The bucket of file in GCS service.                                                                   | -                |
| sources[].gcs.key                           | The object key of file in GCS service.                                                               | -                |
| sources[].gcs.credentialsFile               | Path to the service account or refresh token JSON credentials file. Not required for public data.    | -                |
| sources[].gcs.credentialsJSON               | Content of the service account or refresh token JSON credentials file. Not required for public data. | -                |
| sources[].batch                             | Specifies the batch size for this source of the inserted data.                                       | -                |
| sources[].csv                               | Describes the csv file format information.                                                           | -                |
| sources[].csv.delimiter                     | Specifies the delimiter for the CSV files.                                                           | ","              |
| sources[].csv.withHeader                    | Specifies whether to ignore the first record in csv file.                                            | false            |
| sources[].csv.lazyQuotes                    | Specifies lazy quotes of csv file.                                                                   | false            |
| sources[].csv.comment                       | Specifies the comment character.                                                                     | -                |
| sources[].tags                              | Describes the schema definition for tags.                                                            | -                |
| sources[].tags[].name                       | The tag name.                                                                                        | -                |
| sources[].tags[].mode                       | The mode for processing data, one of `INSERT`, `UPDATE` or `DELETE`.                                 | -                |
| sources[].tags[].filter                     | The data filtering configuration.                                                                    | -                |
| sources[].tags[].filter.expr                | The filter expression.                                                                               | -                |
| sources[].tags[].id                         | Describes the tag ID information.                                                                    | -                |
| sources[].tags[].id.type                    | The type for ID                                                                                      | "STRING"         |
| sources[].tags[].id.index                   | The column number in the records.                                                                    | -                |
| sources[].tags[].id.concatItems             | The concat items to generate for IDs.                                                                | -                |
| sources[].tags[].id.function                | Function to generate the IDs.                                                                        | -                |
| sources[].tags[].ignoreExistedIndex         | Specifies whether to enable `IGNORE_EXISTED_INDEX`.                                                  | true             |
| sources[].tags[].props                      | Describes the tag props definition.                                                                  | -                |
| sources[].tags[].props[].name               | The property name, must be the same with the tag property in NebulaGraph.                            | -                |
| sources[].tags[].props[].type               | The property type.                                                                                   | -                |
| sources[].tags[].props[].index              | The column number in the records.                                                                    | -                |
| sources[].tags[].props[].nullable           | Whether this prop property can be `NULL`.                                                            | false            |
| sources[].tags[].props[].nullValue          | The value used to determine whether it is a `NULL`.                                                  | ""               |
| sources[].tags[].props[].alternativeIndices | The alternative indices.                                                                             | -                |
| sources[].tags[].props[].defaultValue       | The property default value.                                                                          | -                |
| sources[].edges                             | Describes the schema definition for edges.                                                           | -                |
| sources[].edges[].name                      | The edge name.                                                                                       | -                |
| sources[].tags[].mode                       | The `mode` here is similar to `mode` in the `tags` above.                                            | -                |
| sources[].tags[].filter                     | The `filter` here is similar to `filter` in the `tags` above.                                        | -                |
| sources[].edges[].src                       | Describes the source definition for the edge.                                                        | -                |
| sources[].edges[].src.id                    | The `id` here is similar to `id` in the `tags` above.                                                | -                |
| sources[].edges[].dst                       | Describes the destination definition for the edge.                                                   | -                |
| sources[].edges[].dst.id                    | The `id` here is similar to `id` in the `tags` above.                                                | -                |
| sources[].edges[].rank                      | Describes the rank definition for the edge.                                                          | -                |
| sources[].edges[].rank.index                | The column number in the records.                                                                    | -                |
| sources[].edges[].props                     | Similar to the `props` in the `tags`, but for edges.                                                 | -                |
