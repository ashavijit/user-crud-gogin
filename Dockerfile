FROM golang:1.12.5-alpine3.9 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.9

COPY --from=build /app/app /app/app

CMD ["/app/app"]

EXPOSE 8080