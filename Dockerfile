FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

WORKDIR /app

RUN cp /build/main .

FROM scratch

COPY --from=builder /app/main /

EXPOSE 9000

ENTRYPOINT ["/main"]