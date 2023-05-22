FROM reg.vesoft-inc.com/proxy/library/alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata

ADD nebula-importer /usr/local/bin/nebula-importer

ENTRYPOINT ["/usr/local/bin/nebula-importer"]