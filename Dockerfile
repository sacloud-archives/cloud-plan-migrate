FROM golang:1.11 AS builder
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN  apt-get update && apt-get -y install \
        bash \
        git  \
        make \
      && apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

ADD . /go/src/github.com/sacloud/cloud-plan-migrate
WORKDIR /go/src/github.com/sacloud/cloud-plan-migrate
RUN ["make", "tools", "clean", "build"]

#----------

FROM alpine:3.7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

WORKDIR /work

RUN set -x && apk add --no-cache --update ca-certificates
COPY --from=builder /go/src/github.com/sacloud/cloud-plan-migrate/bin/cloud-plan-migrate /usr/local/bin/
RUN chmod +x /usr/local/bin/cloud-plan-migrate
ENTRYPOINT ["/usr/local/bin/cloud-plan-migrate"]
