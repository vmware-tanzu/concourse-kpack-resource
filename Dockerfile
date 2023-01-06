ARG GOLANG_IMG=golang:1.18
FROM ${GOLANG_IMG} as builder

COPY . /code

WORKDIR /code

RUN unset GOPATH && \
    go install ./...

FROM gcr.io/paketo-buildpacks/run-bionic-base:latest

USER root

RUN mkdir -p /opt/resource

COPY --from=builder /root/go/bin/resource /opt/resource/resource

RUN ln -s /opt/resource/resource /opt/resource/check \
    && ln -s /opt/resource/resource /opt/resource/in \
    && ln -s /opt/resource/resource /opt/resource/out
