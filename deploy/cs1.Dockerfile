# syntax=docker/dockerfile:1

FROM golang:1.18-alpine AS build

WORKDIR /ganache

COPY . /ganache

WORKDIR /ganache/cacheserver

# RUN go mod tidy && go mod vendor
RUN CGO_ENABLED=0 go build -o /bin/cacheserver ./cmd/server

FROM scratch
COPY --from=build /bin/cacheserver /bin/cacheserver
COPY --from=build /ganache/certs /ganache/certs

ENTRYPOINT [ "/bin/cacheserver", "-port", "44443" ]
EXPOSE 44443