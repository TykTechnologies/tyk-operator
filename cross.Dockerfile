FROM gcr.io/distroless/static:nonroot
WORKDIR /dist
WORKDIR /
COPY manager.linux manager
USER 65532:65532

ENTRYPOINT ["/manager"]
