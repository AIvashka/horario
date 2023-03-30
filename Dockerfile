FROM golang:1.19-alpine3.15 AS build

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/main.go

FROM alpine:3.15
WORKDIR /app
COPY --from=build /app/main ./

CMD ["./main"]
