FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go build -o /out/polycloud-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates rclone
COPY --from=build /out/polycloud-api /usr/local/bin/polycloud-api
EXPOSE 8080
ENTRYPOINT ["polycloud-api"]
