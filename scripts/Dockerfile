# 名称：nantianmen-ai-gateway通用镜像
# 固定基础镜像版本，避免 latest 漂移导致不可复现的构建
FROM alpine:3.20

RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.aliyun.com/alpine|g' /etc/apk/repositories  \
    && apk update \
    && apk add --no-cache curl tzdata bind-tools

ENV TZ=Asia/Shanghai

RUN ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

ARG APP

ENV NSQ_ADDR=${APP}-nsq:4150
ENV NSQ_TOPIC_PREFIX=${APP}

RUN mkdir -p /${APP}

COPY cmd/* /${APP}/
COPY resource/* /${APP}/

WORKDIR /$APP
# 安全说明：不再内置默认 ADMIN_PASSWORD。
# 部署时必须显式注入管理员初始密码（docker run -e ADMIN_PASSWORD=... 或 compose
# environment），镜像启动脚本在缺失时会明确报错而不是使用弱默认值。
CMD ./docker_run.sh