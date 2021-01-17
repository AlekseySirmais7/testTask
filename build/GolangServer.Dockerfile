FROM golang:latest AS build
RUN go get -v -u 'github.com/lib/pq'
RUN go get -v -u 'github.com/valyala/fasthttp'
RUN go get -v -u 'github.com/qiangxue/fasthttp-routing'
RUN go get -v -u 'github.com/mailru/easyjson/...'
RUN go get -v -u 'github.com/tealeg/xlsx'
RUN go get -v -u 'github.com/pkg/errors'
RUN go get -v -u 'go.uber.org/zap'
RUN go get -v -u 'github.com/golang/mock/gomock'
RUN go get -v -u 'github.com/tarantool/go-tarantool'


WORKDIR /go/src/testTask

COPY . .

RUN GOPATH=/go CGO_ENABLED=0  GOOS=linux  go build -o /main   ./cmd/main

FROM ubuntu:18.04 AS release
MAINTAINER Alex Sirmais

WORKDIR /app
COPY --from=build /main .
RUN chmod +x ./main

EXPOSE 8080/tcp

USER root
CMD sleep 3 && ./main