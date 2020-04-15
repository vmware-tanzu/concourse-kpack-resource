FROM golang:1.14 as builder

COPY . /code

WORKDIR /code

RUN unset GOPATH && \
    go install ./...

FROM cloudfoundry/run:base

RUN mkdir -p /opt/resource

COPY --from=builder /root/go/bin/resource /opt/resource/resource

RUN ln -s /opt/resource/resource /opt/resource/check \
    && ln -s /opt/resource/resource /opt/resource/in \
    && ln -s /opt/resource/resource /opt/resource/out
