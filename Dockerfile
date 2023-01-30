# Build image (Golang)
FROM golang:alpine AS build-stage
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN apk add --no-cache gcc git make

WORKDIR /src
ADD . .

RUN go mod download
RUN go build -ldflags="-s -w" -o cercat ./cmd/

# Final Docker image
FROM chromedp/headless-shell:latest
LABEL MAINTAINER "Thomas Labarussias <issif+github@gadz.org>"

RUN export DEBIAN_FRONTEND=noninteractive \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    dumb-init ca-certificates \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Create user cercat
RUN groupadd --gid 1234 cercat && adduser --uid 1234 --gid 1234 --shell /bin/sh cercat
USER 1234

WORKDIR ${HOME}/
COPY --from=build-stage /src/cercat .

EXPOSE 6060

ENTRYPOINT ["./cercat"]