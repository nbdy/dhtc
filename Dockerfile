FROM golang:latest

WORKDIR /dhtc

COPY cache /dhtc/cache
COPY config /dhtc/config
COPY db /dhtc/db
COPY dhtc-client /dhtc/dhtc-client
COPY notifier /dhtc/notifier
COPY ui /dhtc/ui
COPY go.mod /dhtc/go.mod
COPY go.sum /dhtc/go.sum
COPY main.go /dhtc/main.go

RUN go build

EXPOSE 4200

CMD ["/dhtc/dhtc"]
