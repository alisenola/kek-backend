FROM golang:1.17-alpine AS build

RUN mkdir -p /go/src/github.com/hirokimoto/kek-server-v2 ~/.ssh && \
    apk add --no-cache git openssh-client make gcc libc-dev
WORKDIR /go/src/github.com/hirokimoto/kek-server-v2
COPY . .
RUN make build

FROM alpine:3
COPY --from=build /go/src/github.com/hirokimoto/kek-server-v2/kek-server /bin/kek-server
CMD /bin/kek-server