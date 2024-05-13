FROM golang:latest

# RUN apk add git
WORKDIR /workdir
RUN git config --global --add safe.directory /workdir

COPY scripts ./scripts
COPY cli /cli

ENTRYPOINT ["/cli"]
