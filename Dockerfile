FROM golang:latest

WORKDIR /dhtc

COPY static/ ./static
COPY templates/ ./templates/
COPY go.mod ./
COPY go.sum ./
COPY main.go ./

RUN go build

EXPOSE 4200

CMD ["/dhtc/dhtc"]
