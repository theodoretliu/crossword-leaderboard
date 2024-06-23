FROM golang:1.21.11

RUN apt update && apt upgrade -y

RUN mkdir /app
COPY go.mod go.sum /app
WORKDIR /app
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN go build -o ./server.bin ./cmd/server
RUN go build -o ./scraper.bin ./cmd/scraper
RUN go build -o ./scrapeall.bin ./cmd/scrapeall

FROM ubuntu
RUN apt update && apt upgrade -y && apt install -y ca-certificates
COPY --from=0 /app/server.bin /app/scraper.bin /app/scrapeall.bin /bin/
ENTRYPOINT /bin/server.bin
