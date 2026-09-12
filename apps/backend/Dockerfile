FROM golang:1.25-alpine AS build
WORKDIR /src
# Unduh dependensi lebih dulu supaya layer ini ter-cache saat kode berubah.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/polycloud-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/polycloud-api /usr/local/bin/polycloud-api
EXPOSE 8080
ENTRYPOINT ["polycloud-api"]
