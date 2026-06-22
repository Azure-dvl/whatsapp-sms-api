FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -v -o /usr/local/bin/app 

FROM alpine:latest
WORKDIR /app
COPY --from=builder /usr/local/bin/app /app
ENV DATABASE_URL='postgres://username:password@localhost:5432/database_name'
EXPOSE 9050
CMD ["./app"]