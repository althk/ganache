# syntax=docker/dockerfile:1

FROM golang:1.18-alpine AS build

WORKDIR /ganache

COPY . /ganache

WORKDIR /ganache/cfe

# RUN go mod tidy && go mod vendor
RUN CGO_ENABLED=0 go build -o /bin/cfe ./cmd/server

FROM scratch
COPY --from=build /bin/cfe /bin/cfe
COPY --from=build /ganache/certs /ganache/certs

ENTRYPOINT [ "/bin/cfe", "-port", "40001" ]
EXPOSE 40001