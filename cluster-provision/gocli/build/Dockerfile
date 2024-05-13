FROM alpine:3.18.0

RUN apk add git
WORKDIR /workdir
RUN git config --global --add safe.directory /workdir

COPY cmd/scripts ./scripts
COPY cli /cli

ENTRYPOINT ["/cli"]
