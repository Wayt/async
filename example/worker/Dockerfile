FROM golang:alpine AS build

RUN mkdir -p /go/src/github.com/wayt/async
ADD ../../ /go/src/github.com/wayt/async/

RUN cd /go/src/github.com/wayt/async/examples/client && \
    go build -o client

FROM alpine
COPY --from=build /go/src/github.com/wayt/async/examples/client/client /usr/bin/client

EXPOSE 8080
ENTRYPOINT /usr/bin/client
