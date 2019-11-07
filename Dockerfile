FROM golang:1.13.2-alpine as builder

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn
ENV NEBULA_IMPORTER /home/nebula-importer

COPY . ${NEBULA_IMPORTER}

WORKDIR ${NEBULA_IMPORTER}

RUN cd cmd \
  && go build -o target/nebula-importer . \
  && cp target/nebula-importer /usr/local/bin/nebula-importer

FROM alpine

COPY --from=builder /usr/local/bin/nebula-importer /usr/local/bin/nebula-importer

WORKDIR /root

ENTRYPOINT ["nebula-importer"]
