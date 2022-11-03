FROM golang:1.17 AS builder

WORKDIR /app
COPY . .
ENV CGO_ENABLED 0
RUN make build

FROM gcr.io/distroless/static:debug

COPY --from=builder /app/gojq /
ENTRYPOINT ["/gojq"]
CMD ["--help"]
