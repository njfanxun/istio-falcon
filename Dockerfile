FROM golang:1.18-alpine as GO_BUILD
ARG VERSION
ARG BUILD_DATE
ARG GIT_COMMIT
ARG GIT_TREE_STATE
ARG GIT_TAG
ENV TIMEZONE Asia/Shanghai
ENV GOARCH amd64
ENV GOOS linux
ENV GOFLAGS -buildvcs=false
ENV LDFLAG "-X github.com/njfanxun/istio-falcon/pkg/version.gitVersion=$VERSION \
           -X github.com/njfanxun/istio-falcon/pkg/version.buildDate=$BUILD_DATE \
           -X github.com/njfanxun/istio-falcon/pkg/version.gitCommit=$GIT_COMMIT \
           -X github.com/njfanxun/istio-falcon/pkg/version.gitTreeState=$GIT_TREE_STATE \
           -X github.com/njfanxun/istio-falcon/pkg/version.gitTag=$GIT_TAG"

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
WORKDIR /go/src
ADD . /go/src

RUN cd /go/src \
    && go env -w GOPROXY=https://goproxy.cn \
    && go get -d -v \
    && go install -v
RUN GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$LDFLAG" -o falcon ./main.go

FROM alpine:3.14
LABEL maintainer="fanxun<67831061@qq.com>"
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add -U tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY --from=GO_BUILD /go/src/falcon /app/
ENTRYPOINT ["./falcon"]