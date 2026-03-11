FROM golang:1.25-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=builder /src/server /app/server
COPY --from=builder /src/data /app/data

EXPOSE 9090
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
