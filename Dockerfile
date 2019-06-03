# build stage
FROM golang:1.12 as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# final stage
FROM alpine
COPY --from=builder /app/terracognita /app/
# https://github.com/hashicorp/terraform/issues/10779
RUN apk --update add ca-certificates
ENTRYPOINT ["/app/terracognita"]
