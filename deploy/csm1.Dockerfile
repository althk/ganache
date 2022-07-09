# syntax=docker/dockerfile:1

FROM golang:1.18-alpine AS build

WORKDIR /ganache

COPY . /ganache

WORKDIR /ganache/csm

# RUN go mod tidy && go mod vendor
RUN CGO_ENABLED=0 go build -o /bin/csm ./cmd/server

FROM scratch
COPY --from=build /bin/csm /bin/csm
COPY --from=build /ganache/certs /ganache/certs

ENTRYPOINT [ "/bin/csm", "-port", "41443" ]
EXPOSE 41443