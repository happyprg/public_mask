# Simple usage with a mounted data directory:
FROM golang:1.13-alpine AS build-env
ENV PACKAGES make perl
RUN apk add --no-cache $PACKAGES

# Set working directory for the build
WORKDIR /work_dir
#RUN go mod download
# Add source files
COPY . .

RUN go build -o ./public_mask

# Final image
FROM alpine:edge
ENV PACKAGES curl make git perl
RUN apk add --no-cache $PACKAGES
COPY --from=build-env /work_dir/public_mask /public_mask
CMD ./public_mask

