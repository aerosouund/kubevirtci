FROM golang:1.22

# RUN apk add git
WORKDIR /workdir
RUN git config --global --add safe.directory /workdir

COPY scripts ./scripts
COPY manifests ./manifests
COPY cli /cli

ENTRYPOINT ["/cli"]
