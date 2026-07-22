# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY app/ ./

RUN go build -o manager ./cmd/manager

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /server

RUN apt-get update && \
    apt-get install -y \
    curl \
    wget \
    bash \
    tar \
    && rm -rf /var/lib/apt/lists/*


# Install Java JRE versions
RUN mkdir -p /opt/java && \
    \
    # Java 8 JRE
    wget -q https://api.adoptium.net/v3/binary/latest/8/ga/linux/x64/jre/hotspot/normal/eclipse \
        -O /tmp/java8.tar.gz && \
    mkdir -p /opt/java/8 && \
    tar -xzf /tmp/java8.tar.gz -C /opt/java/8 --strip-components=1 && \
    \
    # Java 17 JRE
    wget -q https://api.adoptium.net/v3/binary/latest/17/ga/linux/x64/jre/hotspot/normal/eclipse \
        -O /tmp/java17.tar.gz && \
    mkdir -p /opt/java/17 && \
    tar -xzf /tmp/java17.tar.gz -C /opt/java/17 --strip-components=1 && \
    \
    # Java 21 JRE
    wget -q https://api.adoptium.net/v3/binary/latest/21/ga/linux/x64/jre/hotspot/normal/eclipse \
        -O /tmp/java21.tar.gz && \
    mkdir -p /opt/java/21 && \
    tar -xzf /tmp/java21.tar.gz -C /opt/java/21 --strip-components=1 && \
    \
    rm -f /tmp/*.tar.gz


# Default Java
ENV JAVA_HOME=/opt/java/21
ENV PATH="${JAVA_HOME}/bin:${PATH}"

COPY --from=builder /app/manager /usr/local/bin/manager

COPY --from=builder /app/server-test /server/server-test

CMD ["tail", "-f", "/dev/null"]