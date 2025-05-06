# Build meter in a stock Go builder container
FROM meterio/build-env:24.04 AS builder

WORKDIR  /app

COPY . .
ENV GOROOT=/usr/local/go
RUN make

# Pull meter into a second stage deploy alpine container
FROM ubuntu:24.04

# RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bitcoind_exporter /usr/bin/
ENV LD_LIBRARY_PATH=/usr/lib
ENV BTC_USER=testuser
ENV BTC_PASS=testpass
ENV HTTP_LISTENADDR=:8333

EXPOSE 8333 
ENTRYPOINT ["bitcoind_exporter"]
