FROM node:20 AS css-builder
WORKDIR /build
COPY package*.json ./
RUN npm install
COPY ui ./ui
COPY tailwind.config.js ./
RUN npm run build:css

FROM golang:latest AS go-builder
WORKDIR /dhtc
COPY . .
COPY --from=css-builder /build/ui/static/css/style.css ./ui/static/css/style.css
RUN go build -o dhtc cmd/dhtc/main.go

FROM debian:bookworm-slim
WORKDIR /dhtc
COPY --from=go-builder /dhtc/dhtc .
EXPOSE 4200
CMD ["/dhtc/dhtc"]
