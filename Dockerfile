FROM golang:alpine AS builder

ADD ./ /go/src/github.com/tamslo/gitlab-issue-automation

RUN set -ex && \
  cd /go/src/github.com/tamslo/gitlab-issue-automation && \       
  CGO_ENABLED=0 go build \
        -tags netgo \
        -v -a \
        -ldflags '-extldflags "-static"' \
        -buildvcs=false  && \
  mv ./gitlab-issue-automation /usr/bin/gitlab-issue-automation

FROM busybox

COPY --from=builder /usr/bin/gitlab-issue-automation /usr/local/bin/gitlab-issue-automation

ENTRYPOINT [ "gitlab-issue-automation" ]