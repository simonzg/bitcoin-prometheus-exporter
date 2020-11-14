# Build meter in a stock Go builder container
FROM dfinlab/build-env as builder

WORKDIR  /app

COPY . .

RUN make

# Pull meter into a second stage deploy alpine container
FROM ubuntu:18.04

# RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bitcoind_exporter /usr/bin/
ENV LD_LIBRARY_PATH=/usr/lib
ENV BTC_USER=testuser
ENV BTC_PASS=testpass
ENV HTTP_LISTENADDR=:8333

EXPOSE 8333 
ENTRYPOINT ["bitcoind_exporter"]
