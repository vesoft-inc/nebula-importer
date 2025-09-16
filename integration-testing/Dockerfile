FROM reg.vesoft-inc.com/proxy/library/golang:1.22-alpine

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
WORKDIR /build/zero

RUN go install github.com/vesoft-inc/nebula-console@latest

ADD . .

RUN go mod download
RUN go build -ldflags="-s -w" -o /usr/bin/nebula-importer ./cmd/nebula-importer

ENTRYPOINT ["sh", "-c"]
