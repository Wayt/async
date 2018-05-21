FROM golang:alpine AS build

RUN mkdir -p /go/src/github.com/wayt/async
ADD . /go/src/github.com/wayt/async/

RUN cd /go/src/github.com/wayt/async && \
    go build -o async-bin


FROM alpine
COPY --from=build /go/src/github.com/wayt/async/async-bin /usr/bin/async

# Server port
EXPOSE 8080
EXPOSE 8000

ENTRYPOINT /usr/bin/async server
