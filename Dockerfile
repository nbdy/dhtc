FROM golang:latest

WORKDIR /dhtc

COPY cache /dhtc/cache
COPY config /dhtc/config
COPY db /dhtc/db
COPY dhtc-client /dhtc/dhtc-client
COPY notifier /dhtc/notifier
COPY ui /dhtc/ui
COPY cmd /dhtc/cmd
COPY go.mod /dhtc/go.mod
COPY go.sum /dhtc/go.sum

RUN go build -o dhtc /dhtc/cmd/dhtc/main.go

EXPOSE 4200

CMD ["/dhtc/dhtc"]
