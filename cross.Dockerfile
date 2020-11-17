FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY manager.linux manager
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
