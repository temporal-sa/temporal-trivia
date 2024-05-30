FROM docker.io/library/golang:1.20-alpine
RUN mkdir -p /app/bin
WORKDIR /app
ENV GOBIN=/app/bin
COPY . .

RUN go install worker/worker.go

FROM docker.io/alpine:latest
LABEL ios.k8s.display-name="backup-worker" \
    maintainer="Keith Tenzer <keith.tenzer@temporal.io>"

RUN mkdir -p /app/bin
WORKDIR /app/bin
COPY --from=0 /app/bin/worker /app/bin
CMD ["/app/bin/worker" ]
