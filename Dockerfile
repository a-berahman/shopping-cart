FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk update && apk add --no-cache git
RUN apk add --no-cache git ca-certificates tzdata gcc g++ make musl-dev curl


RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate /usr/local/bin/migrate
CMD ["sh", "-c", "migrate -path=/app/migrations -database=${DATABASE_URL} up && ./app"]


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN cp config.yaml.example config.yaml
RUN go test -v ./...
RUN CGO_ENABLED=0 go build -o /go/bin/shopping-cart ./cmd/main.go


FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -g '' appuser

COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /go/bin/shopping-cart .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/migrations ./migrations


ENV CONFIG_PATH=/config.yaml
ENV TZ=UTC

EXPOSE 8080

CMD ["/shopping-cart", "--config-path", "config.yaml"]