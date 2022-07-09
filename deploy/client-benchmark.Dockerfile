# syntax=docker/dockerfile:1

FROM golang:1.18-alpine AS build

WORKDIR /ganache

COPY . /ganache

WORKDIR /ganache/client

RUN CGO_ENABLED=0 go test -c -o /bin/benchmark_test .

FROM scratch
COPY --from=build /bin/benchmark_test /bin/benchmark_test
COPY --from=build /ganache/certs /ganache/certs

CMD [ "/bin/benchmark_test", "-test.bench=CFE", "-test.benchmem", "-test.benchtime=10000x", "-cfe_server=ganache_cfe_1:40001", "-root_ca_file=/ganache/certs/testca.crt" ]