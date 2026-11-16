FROM alpine:3
RUN apk add --no-cache ca-certificates
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/markdown2confluence /markdown2confluence
ENTRYPOINT [ "/markdown2confluence" ]
