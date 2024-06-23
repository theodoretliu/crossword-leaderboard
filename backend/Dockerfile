FROM golang:1.21.4

RUN apt update && apt upgrade -y

RUN mkdir /app
COPY *.go *.yml go.mod go.sum /app
WORKDIR /app
RUN go get
RUN go build

ENTRYPOINT ./server
