FROM golang:1.16.4-alpine3.13 as GO_BUILD
ENV GOARCH amd64
ENV GOOS linux
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
WORKDIR /go/src
ADD . /go/src
RUN cd /go/src \
    && go env -w GOPROXY=https://goproxy.cn \
    && go get -d -v \
    && go install -v
RUN GOOS=$GOOS GOARCH=$GOARCH go build -v -i -o istio-falcon ./main.go
FROM alpine:3.13
LABEL maintainer="fanxun<67831061@qq.com>"
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add -U tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY --from=GO_BUILD /go/src/istio-falcon /app/
ENTRYPOINT ["./istio-falcon"]