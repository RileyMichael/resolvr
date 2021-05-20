FROM golang:1.16 as builder

RUN mkdir /build
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o resolvr ./cmd/resolvr

FROM alpine:latest

RUN mkdir /app
WORKDIR /app

COPY --from=builder /build/resolvr .

ENTRYPOINT ["./resolvr"]
