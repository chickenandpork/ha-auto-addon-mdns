FROM golang:1.24-bookworm AS builder

# Docker Buildx sets these automatically for multi-platform builds
ARG TARGETARCH
ARG TARGETVARIANT

ENV GOARCH=$TARGETARCH
RUN case "$TARGETARCH" in \
        arm)  export GOARM=$(echo $TARGETVARIANT | tr -d 'v') ;; \
        amd64) export GOAMD64=auto ;; \
    esac

# Install Bazelisk
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl ca-certificates git && \
    curl -L https://github.com/bazelbuild/bazelisk/releases/download/v1.23.0/bazelisk-linux-${TARGETARCH} -o /usr/local/bin/bazel && \
    chmod +x /usr/local/bin/bazel

# Copy source
COPY . /src
WORKDIR /src

# Build the binary (cross-compiled via GOARCH)
RUN bazel build //cmd/auto-mdns:auto-mdns && \
    cp bazel-bin/cmd/auto-mdns/auto-mdns_/auto-mdns /usr/local/bin/auto-mdns

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl jq && \
    rm -rf /var/lib/apt/lists/*

COPY addon/start.sh /usr/local/bin/start.sh
RUN chmod +x /usr/local/bin/start.sh

COPY --from=builder /usr/local/bin/auto-mdns /usr/local/bin/auto-mdns

ENV SUPERVISOR_URL=http://supervisor
ENV POLL_INTERVAL=30s

EXPOSE 80
EXPOSE 443

CMD ["/usr/local/bin/start.sh"]
