[![codecov.io](https://codecov.io/gh/vesoft-inc/nebula-importer/branch/master/graph/badge.svg)](https://codecov.io/gh/vesoft-inc/nebula-importer)
[![Go Report Card](https://goreportcard.com/badge/github.com/vesoft-inc/nebula-importer)](https://goreportcard.com/report/github.com/vesoft-inc/nebula-importer)
[![GolangCI](https://golangci.com/badges/github.com/vesoft-inc/nebula-importer.svg)](https://golangci.com/r/github.com/vesoft-inc/nebula-importer)
[![GoDoc](https://godoc.org/github.com/vesoft-inc/nebula-importer?status.svg)](https://godoc.org/github.com/vesoft-inc/nebula-importer)

# What is NebulaGraph Importer?

**NebulaGraph Importer** is a tool to import data into [NebulaGraph](https://github.com/vesoft-inc/nebula).

## Features

* Support multiple data sources, currently supports `local`, `s3`, `oss`, `ftp`, `sftp`, `hdfs`, and `gcs`.
* Support multiple file formats, currently only `csv` files are supported.
* Support files containing multiple tags, multiple edges, and a mixture of both.
* Support data transformations.
* Support record filtering.
* Support multiple modes, including `INSERT`, `UPDATE`, `DELETE`.
* Support connect multiple Graph with automatically load balance.
* Support retry after failure.
* Humanized status printing.

_See configuration instructions for more features._

## How to Install

### From Releases

Download the packages on the [Releases page](https://github.com/vesoft-inc/nebula-importer/releases), and give execute permissions to it.

You can choose according to your needs, the following installation packages are supported:

* binary
* archives
* apk
* deb
* rpm

### From go install

```shell
$ go install github.com/vesoft-inc/nebula-importer/cmd/nebula-importer@latest
```

### From docker

```shell
$ docker pull vesoft/nebula-importer:<version>
$ docker run --rm -ti \
      --network=host \
      -v <config_file>:<config_file> \
      -v <data_dir>:<data_dir> \
      vesoft/nebula-importer:<version>
      --config <config_file>

# config_file: the absolute path to the configuration file.
# data_dir: the absolute path to the data directory, ignore if not a local file.
# version: the version of NebulaGraph Importer.
```

### From Source Code

```shell
$ git clone https://github.com/vesoft-inc/nebula-importer
$ cd nebula-importer
$ make build
```

You can find a binary named `nebula-importer` in `bin` directory.

## Configuration Instructions

`NebulaGraph Importer`'s configuration file is in YAML format. You can find some examples in [examples](examples/).

Configuration options are divided into four groups:

* `client` is configuration options related to the NebulaGraph connection client.
* `manager` is global control configuration options related to NebulaGraph Importer.
* `log` is configuration options related to printing logs.
* `sources` is the data source configuration items.

### client

```yaml
client:
  version: v3
  address: "127.0.0.1:9669"
  user: root
  password: nebula
  ssl:
    enable: true
    certPath: "your/cert/file/path"
    keyPath: "your/key/file/path"
    caPath: "your/ca/file/path"
    insecureSkipVerify: false
  concurrencyPerAddress: 16
  reconnectInitialInterval: 1s
  retry: 3
  retryInitialInterval: 1s
```

* `client.version`: **Required**. Specifies which version of NebulaGraph, currently only `v3` is supported.
* `client.address`: **Required**. The address of graph in NebulaGraph.
* `client.user`: **Optional**. The user of NebulaGraph. The default value is `root`.
* `client.password`: **Optional**. The password of NebulaGraph. The default value is `nebula`.
* `client.ssl`: **Optional**. SSL related configuration.
* `client.ssl.enable`: **Optional**. Specifies whether to enable ssl authentication. The default value is `false`.
* `client.ssl.certPath`: **Required**. Specifies the path of the certificate file.
* `client.ssl.keyPath`: **Required**. Specifies the path of the private key file.
* `client.ssl.caPath`: **Required**. Specifies the path of the certification authority file.
* `client.ssl.insecureSkipVerify`: **Optional**. Specifies whether a client verifies the server's certificate chain and host name. The default value is `false`.
* `client.concurrencyPerAddress`: **Optional**. The number of client connections to each graph in NebulaGraph. The default value is `10`.
* `client.reconnectInitialInterval`: **Optional**. The initialization interval for reconnecting NebulaGraph. The default value is `1s`.
* `client.retry`: **Optional**. The failed retrying times to execute nGQL queries in NebulaGraph client. The default value is `3`.
* `client.retryInitialInterval`: **Optional**. The initialization interval retrying. The default value is `1s`.

### manager

```yaml
  spaceName: basic_int_examples
  batch: 128
  readerConcurrency: 50
  importerConcurrency: 512
  statsInterval: 10s
  hooks:
    before:
      - statements:
          - UPDATE CONFIGS storage:wal_ttl=3600;
          - UPDATE CONFIGS storage:rocksdb_column_family_options = { disable_auto_compactions = true };
      - statements:
          - |
            DROP SPACE IF EXISTS basic_int_examples;
            CREATE SPACE IF NOT EXISTS basic_int_examples(partition_num=5, replica_factor=1, vid_type=int);
            USE basic_int_examples;
        wait: 10s
    after:
      - statements:
          - |
            UPDATE CONFIGS storage:wal_ttl=86400;
            UPDATE CONFIGS storage:rocksdb_column_family_options = { disable_auto_compactions = false };
```

* `manager.spaceName`: **Required**. Specifies which space the data is imported into.
* `manager.batch`: **Optional**. Specifies the batch size for all sources of the inserted data. The default value is `128`.
* `manager.readerConcurrency`: **Optional**. Specifies the concurrency of reader to read from sources. The default value is `50`.
* `manager.importerConcurrency`: **Optional**. Specifies the concurrency of generating inserted nGQL statement, and then call client to import. The default value is `512`.
* `manager.statsInterval`: **Optional**. Specifies the interval at which statistics are printed. The default value is `10s`.
* `manager.hooks.before`: **Optional**. Configures the statements before the import begins.
  * `manager.hooks.before.[].statements`: Defines the list of statements.
  * `manager.hooks.before.[].wait`: **Optional**. Defines the waiting time after executing the above statements.
* `manager.hooks.after`: **Optional**. Configures the statements after the import is complete.
  * `manager.hooks.after.[].statements`: **Optional**. Defines the list of statements.
  * `manager.hooks.after.[].wait`: **Optional**. Defines the waiting time after executing the above statements.

### log

```yaml
log:
  level: INFO
  console: true
  files:
    - logs/nebula-importer.log
```

* `log.level`: **Optional**. Specifies the log level, optional values is `DEBUG`, `INFO`, `WARN`, `ERROR`, `PANIC` or `FATAL`. The default value is `INFO`.
* `log.console`: **Optional**. Specifies whether to print logs to the console. The default value is `true`.
* `log.files`: **Optional**. Specifies which files to print logs to.

### sources

`sources` is the configuration of the data source list, each data source contains data source information, data processing and schema mapping.

The following are the relevant configuration items.

* `batch` specifies the batch size for this source of the inserted data. The priority is greater than `manager.batch`.
* `path`, `s3`, `oss`, `ftp`, `sftp`, `hdfs`, and `gcs` are information configurations of various data sources, and only one of them can be configured.
* `csv` describes the csv file format information.
* `tags` describes the schema definition for tags.
* `edges` describes the schema definition for edges.

#### path

It only needs to be configured for local file data sources.

```yaml
path: ./person.csv
```

* `path`: **Required**. Specifies the path where the data files are stored. If a relative path is used, the path and current configuration file directory are spliced. Wildcard filename is also supported, for example: ./follower-*.csv, please make sure that all matching files with the same schema.

#### s3

It only needs to be configured for s3 data sources.

```yaml
s3:
  endpoint: <endpoint>
  region: <region>
  bucket: <bucket>
  key: <key>
  accessKeyID: <Access Key ID>
  accessKeySecret: <Access Key Secret>
```

* `endpoint`: **Optional**. The endpoint of s3 service, can be omitted if using aws s3.
* `region`: **Required**. The region of s3 service.
* `bucket`: **Required**. The bucket of file in s3 service.
* `key`: **Required**. The object key of file in s3 service.
* `accessKeyID`: **Optional**. The `Access Key ID` of s3 service. If it is public data, no need to configure.
* `accessKeySecret`: **Optional**. The `Access Key Secret` of s3 service. If it is public data, no need to configure.

#### oss

It only needs to be configured for oss data sources.

```yaml
oss:
  endpoint: <endpoint>
  bucket: <bucket>
  key: <key>
  accessKeyID: <Access Key ID>
  accessKeySecret: <Access Key Secret>
```

* `endpoint`: **Required**. The endpoint of oss service.
* `bucket`: **Required**. The bucket of file in oss service.
* `key`: **Required**. The object key of file in oss service.
* `accessKeyID`: **Required**. The `Access Key ID` of oss service.
* `accessKeySecret`: **Required**. The `Access Key Secret` of oss service.

#### ftp

It only needs to be configured for ftp data sources.

```yaml
ftp:
  host: 192.168.0.10
  port: 21
  user: <user>
  password: <password>
  path: <path of file>
```

* `host`: **Required**. The host of ftp service.
* `port`: **Required**. The port of ftp service.
* `user`: **Required**. The user of ftp service.
* `password`: **Required**. The password of ftp service.
* `path`: **Required**. The path of file in the ftp service.

#### sftp

It only needs to be configured for sftp data sources.

```yaml
sftp:
  host: 192.168.0.10
  port: 22
  user: <user>
  password: <password>
  keyFile: <keyFile>
  keyData: <keyData>
  passphrase: <passphrase>
  path: <path of file>
```

* `host`: **Required**. The host of sftp service.
* `port`: **Required**. The port of sftp service.
* `user`: **Required**. The user of sftp service.
* `password`: **Optional**. The password of sftp service.
* `keyFile`: **Optional**. The ssh key file path of sftp service.
* `keyData`: **Optional**. The ssh key file content of sftp service.
* `passphrase`: **Optional**. The ssh key passphrase of sftp service.
* `path`: **Required**. The path of file in the sftp service.

#### hdfs

It only needs to be configured for hdfs data sources.

```yaml
hdfs:
  address: 192.168.0.10:8020
  user: <user>
  servicePrincipalName: <Kerberos Service Principal Name>
  krb5ConfigFile: <Kerberos config file>
  ccacheFile: <Kerberos ccache file>
  keyTabFile: <Kerberos keytab file>
  password: <Kerberos password>
  dataTransferProtection: <Kerberos Data Transfer Protection>
  disablePAFXFAST: false
  path: <path of file>
```

* `address`: **Required**. The address of hdfs service.
* `user`: **Optional**. The user of hdfs service.
* `servicePrincipalName`: **Optional**. The kerberos service principal name of hdfs service when enable kerberos.
* `krb5ConfigFile`: **Optional**. The kerberos config file of hdfs service when enable kerberos, default is `/etc/krb5.conf`.
* `ccacheFile`: **Optional**. The ccache file of hdfs service when enable kerberos.
* `keyTabFile`: **Optional**. The keytab file of hdfs service when enable kerberos.
* `password`: **Optional**. The kerberos password of hdfs service when enable kerberos.
* `dataTransferProtection`: **Optional**. The data transfer protection of hdfs service.
* `disablePAFXFAST`: **Optional**. Whether to prohibit the client to use PA_FX_FAST.
* `path`: **Required**. The path of file in the sftp service.

#### gcs

It only needs to be configured for gcs data sources.

```yaml
gcs:
  endpoint: <endpoint>
  bucket: <bucket>
  key: <key>
  credentialsFile: <Service account or refresh token JSON credentials file>
  credentialsJSON: <Service account or refresh token JSON credentials>
  withoutAuthentication: <false | true>
```

* `endpoint`: **Optional**. The endpoint of GCS service.
* `bucket`: **Required**. The bucket of file in GCS service.
* `key`: **Required**. The object key of file in GCS service.
* `credentialsFile`: **Optional**. Path to the service account or refresh token JSON credentials file. Not required for public data.
* `credentialsJSON`: **Optional**. Content of the service account or refresh token JSON credentials file. Not required for public data.
* `withoutAuthentication`: **Optional**. Specifies that no authentication should be used, defaults to `false`.

#### batch

```yaml
batch: 256
```

* `batch`: **Optional**. Specifies the batch size for this source of the inserted data. The priority is greater than `manager.batch`.

#### csv

```yaml
csv:
  delimiter: ","
  withHeader: false
  lazyQuotes: false
  comment: ""
```

* `delimiter`: **Optional**. Specifies the delimiter for the CSV files. The default value is `","`. And only a 1-character string delimiter is supported.
* `withHeader`: **Optional**. Specifies whether to ignore the first record in csv file. The default value is `false`.
* `lazyQuotes`: **Optional**. If lazyQuotes is true, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field.
* `comment`: **Optional**. Specifies the comment character. Lines beginning with the Comment character without preceding whitespace are ignored.

#### tags

```yaml
tags:
- name: Person
  mode: INSERT
  filter:
    expr: (Record[1] == "Mahinda" or Record[1] == "Michael") and Record[3] == "male"
  id:
    type: "STRING"
    function: "hash"
    index: 0
  ignoreExistedIndex: true
  props:
    - name: "firstName"
      type: "STRING"
      index: 1
    - name: "lastName"
      type: "STRING"
      index: 2
    - name: "gender"
      type: "STRING"
      index: 3
      nullable: true
      defaultValue: male
    - name: "birthday"
      type: "DATE"
      index: 4
      nullable: true
      nullValue: _NULL_
    - name: "creationDate"
      type: "DATETIME"
      index: 5
    - name: "locationIP"
      type: "STRING"
      index: 6
    - name: "browserUsed"
      type: "STRING"
      index: 7
      nullable: true
      alternativeIndices:
        - 6

# concatItems examples
tags:
- name: Person
  id:
    type: "STRING"
    concatItems:
      - "abc"
      - 1
    function: hash
```

* `name`: **Required**. The tag name.
* `mode`: **Optional**. The mode for processing data, optional values is `INSERT`, `UPDATE` or `DELETE`, default `INSERT`.
* `filter`: **Optional**. The data filtering configuration.
  * `expr`: **Required**. The filter expression. See the [Filter Expression](docs/filter-expression.md) for details.
* `id`: **Required**. Describes the tag ID information.
  * `type`: **Optional**. The type for ID. The default value is `STRING`.
  * `index`: **Optional**. The column number in the records. Required if `concatItems` is not configured.
  * `concatItems`: **Optional**. The concat items to generate for IDs. The concat item can be string, int or mixed. string represents a constant, and int represents an index column. Then connect all items. If set, the above index will have no effect.
  * `function`: **Optional**. Functions to generate the IDs. Currently, we only support function `hash`.
* `ignoreExistedIndex`: **Optional**. Specifies whether to enable `IGNORE_EXISTED_INDEX`. The default value is `true`.
* `props`: **Required**. Describes the tag props definition.
  * `name`: **Required**. The property name, must be the same with the tag property in NebulaGraph.
  * `type`: **Optional**. The property type, currently `BOOL`, `INT`, `FLOAT`, `DOUBLE`, `STRING`, `TIME`, `TIMESTAMP`, `DATE`, `DATETIME`, `GEOGRAPHY`, `GEOGRAPHY(POINT)`, `GEOGRAPHY(LINESTRING)` and `geography(polygon)` are supported. The default value is `STRING`.
  * `index`: **Required**. The column number in the records.
  * `nullable`: **Optional**. Whether this prop property can be `NULL`, optional values is `true` or `false`, default `false`.
  * `nullValue`: **Optional**. Ignored when `nullable` is `false`. The value used to determine whether it is a `NULL`. The property is set to `NULL` when the value is equal to `nullValue`, default `""`.
  * `alternativeIndices`: **Optional**. Ignored when `nullable` is `false`. The property is fetched from records according to the indices in order until not equal to `nullValue`.
  * `defaultValue`: **Optional**. Ignored when `nullable` is `false`. The property default value, when all the values obtained by `index` and `alternativeIndices` are `nullValue`.

#### edges

```yaml
edges:
- name: KNOWS
  mode: INSERT
  filter:
    expr: (Record[1] == "Mahinda" or Record[1] == "Michael") and Record[3] == "male"
  src:
    id:
      type: "INT"
      index: 0
  dst:
    id:
      type: "INT"
      index: 1
  rank:
    index: 0
  ignoreExistedIndex: true
  props:
    - name: "creationDate"
      type: "DATETIME"
      index: 2
      nullable: true
      nullValue: _NULL_
      defaultValue: 0000-00-00T00:00:00
```

* `name`: **Required**. The edge name.
* `mode`: **Optional**. The `mode` here is similar to `mode` in the `tags` above.
* `filter`: **Optional**. The `filter` here is similar to `filter` in the `tags` above.
* `src`: **Required**. Describes the source definition for the edge.
* `src.id`: **Required**. The `id` here is similar to `id` in the `tags` above.
* `dst`: **Required**. Describes the destination definition for the edge.
* `dst.id`: **Required**. The `id` here is similar to `id` in the `tags` above.
* `rank`: **Optional**. Describes the rank definition for the edge.
* `rank.index`: **Required**. The column number in the records.
* `props`: **Optional**. Similar to the `props` in the `tags`, but for edges.

See the [Configuration Reference](docs/configuration-reference.md) for details on the configurations.
