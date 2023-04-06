# Build the manager binary
FROM golang:alpine as builder
LABEL maintainer="The Prometheus Authors <shenwei@cmss.chinamobile.com>"

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY hardware_exporter.go hardware_exporter.go
COPY collector/ collector/
COPY lib/       lib/
COPY config.ini config.ini
# Build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=arm64
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn,direct"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build hardware_exporter.go

FROM alpine as hardware_exporter
# 设置时区为上海
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone
# 设置编码
ENV LANG C.UTF-8
WORKDIR /
COPY --from=builder /workspace/hardware_exporter .
COPY --from=builder /workspace/config.ini  .
EXPOSE      9696
ENTRYPOINT ["/hardware_exporter"]
