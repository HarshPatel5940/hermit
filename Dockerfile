# Build stage (golang:1.25-alpine) - builds for the native arch of the builder
FROM golang:1.25-alpine AS build
ARG TARGETARCH

# install build-time utilities; keep minimal
RUN apk add --no-cache curl ca-certificates libstdc++ libgcc git

WORKDIR /app

# cache modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV GOBIN=/usr/local/bin
RUN set -eux; \
    go install github.com/a-h/templ/cmd/templ@v0.3.943; \
    templ --version

RUN set -eux; \
    if [ -n "${TARGETARCH}" ]; then ARCH="${TARGETARCH}"; else \
    UNM="$(uname -m)"; \
    case "${UNM}" in \
    x86_64|amd64) ARCH="amd64" ;; \
    aarch64|arm64) ARCH="arm64" ;; \
    *) ARCH="${UNM}" ;; \
    esac; \
    fi; \
    if [ "${ARCH}" = "arm64" ] || [ "${ARCH}" = "aarch64" ]; then \
    TAILWIND_ASSET="tailwindcss-linux-arm64-musl"; \
    else \
    TAILWIND_ASSET="tailwindcss-linux-x64-musl"; \
    fi; \
    echo "Detected ARCH=${ARCH} TARGETARCH=${TARGETARCH} -> using ${TAILWIND_ASSET}"; \
    curl -fsSL "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/${TAILWIND_ASSET}" -o /usr/local/bin/tailwindcss; \
    chmod +x /usr/local/bin/tailwindcss; \
    # quick sanity check so the build fails with a useful message if the binary can't run
    /usr/local/bin/tailwindcss --version


RUN set -eux; \
    templ generate && \
    tailwindcss -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css

# Build final static binary for linux/native arch
# Set CGO_ENABLED=0 and GOOS=linux so binary is static (no libc dependency).
# Do NOT set GOARCH: leave it to the builder's default (native arch).
ENV CGO_ENABLED=0 \
    GOOS=linux

RUN set -eux; \
    go build -trimpath -ldflags="-s -w" -o /app/main cmd/api/main.go

# Final stage (scratch) - minimal runtime
FROM scratch AS prod

USER 65532

WORKDIR /app
COPY --from=build /app/main /app/main

# expose port (use env in compose)
EXPOSE ${PORT:-8080}
ENV PORT=${PORT:-8080}

CMD ["/app/main"]
