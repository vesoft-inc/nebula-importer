FROM reg.vesoft-inc.com/ci/golang:1.18-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
WORKDIR /build/zero

ADD go.mod .
ADD go.sum .
COPY pkg pkg
COPY cmd cmd
RUN go mod download

RUN go build -ldflags="-s -w" -o /usr/bin/nebula-importer ./cmd/nebula-importer

FROM reg.vesoft-inc.com/ci/alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

COPY --from=builder /usr/bin/nebula-importer /usr/bin/nebula-importer

ENTRYPOINT ["/usr/bin/nebula-importer"]
