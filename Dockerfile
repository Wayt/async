FROM golang:alpine AS build

RUN mkdir -p /go/src/github.com/wayt/async
ADD . /go/src/github.com/wayt/async/

RUN cd /go/src/github.com/wayt/async && \
    go build -o async

FROM alpine
WORKDIR /app
COPY --from=build /go/src/github.com/wayt/async/async /app/

EXPOSE 8080
ENTRYPOINT /app/async server
