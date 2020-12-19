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
FROM alpine AS final-stage
LABEL MAINTAINER "Thomas Labarussias <issif+github@gadz.org>"

RUN apk add --no-cache ca-certificates

# Create user cercat
RUN addgroup -S cercat && adduser -u 1234 -S cercat -G cercat
USER 1234

WORKDIR ${HOME}/
COPY --from=build-stage /src/cercat .

EXPOSE 6060

ENTRYPOINT ["./cercat"]