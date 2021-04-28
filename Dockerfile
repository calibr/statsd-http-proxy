FROM alpine:3.13.2

ARG VERSION

COPY . /src/
WORKDIR /src

RUN apk update && \
    # install requirements
    apk add --no-cache --virtual .build-deps \
        ca-certificates \
        make \
        wget \
        git \
        curl \
        go \
        musl-dev && \
    # update certs
    update-ca-certificates

RUN ls

# make and install source
RUN make build && \
    chmod +x ./bin/statsd-http-proxy && \
    mv ./bin/statsd-http-proxy /usr/local/bin && \
    # clear
    cd .. && rm -rf statsd-http-proxy-${VERSION} && \
    apk del .build-deps


# start service
EXPOSE 80
ENTRYPOINT ["/usr/local/bin/statsd-http-proxy", "--http-host="]