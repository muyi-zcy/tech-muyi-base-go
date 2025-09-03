FROM registry.cn-hangzhou.aliyuncs.com/ideaistudio-hub/golang:1.22.1-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /home/app/ ./main.go


FROM registry.cn-hangzhou.aliyuncs.com/ideaistudio-hub/golang:1.22.1-alpine

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai


WORKDIR /home/app

COPY --from=builder /home/app/ /home/app/
COPY --from=builder /build/ /home/app/

EXPOSE 28080

CMD /bin/sh /home/app/start.sh