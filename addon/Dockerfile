FROM golang:1.24-bookworm AS builder
ARG TARGETARCH
ENV GOARCH=$TARGETARCH

# Install Bazelisk
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl ca-certificates git && \
    curl -L https://github.com/bazelbuild/bazelisk/releases/download/v1.23.0/bazelisk-linux-${TARGETARCH} -o /usr/local/bin/bazel && \
    chmod +x /usr/local/bin/bazel

# Copy source
COPY . /src
WORKDIR /src

# Build the binary
RUN pwd && ls -al && bazel build //cmd/auto-mdns:auto-mdns && \
    cp bazel-bin/cmd/auto-mdns/auto-mdns_/auto-mdns /usr/local/bin/auto-mdns
    # cp $(bazel cquery --output=files //cmd/auto-mdns:auto-mdns 2>/dev/null) /usr/local/bin/auto-mdns

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl jq && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/auto-mdns /usr/local/bin/auto-mdns
COPY addon/start.sh /usr/local/bin/start.sh
RUN chmod +x /usr/local/bin/start.sh

ENV SUPERVISOR_URL=http://supervisor
ENV POLL_INTERVAL=30s

EXPOSE 80
EXPOSE 443

CMD ["/usr/local/bin/start.sh"]
