FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY manager.exe manager
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
