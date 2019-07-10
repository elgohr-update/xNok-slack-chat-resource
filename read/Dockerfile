FROM golang:alpine as build-env
RUN apk update
RUN apk add git
RUN go get -u github.com/golang/dep/cmd/dep
COPY ./utils /go/src/github.com/jleben/slack-chat-resource/utils
COPY Gopkg.toml /go/src/github.com/jleben/slack-chat-resource
RUN ( cd /go/src/github.com/jleben/slack-chat-resource ; dep ensure )
RUN go build -o /assets/utils github.com/jleben/slack-chat-resource/utils

FROM build-env as build-read
COPY ./read/check /go/src/github.com/jleben/slack-chat-resource/read/check/
COPY ./read/in /go/src/github.com/jleben/slack-chat-resource/read/in/
COPY ./read/out /go/src/github.com/jleben/slack-chat-resource/read/out/
RUN go build -o /assets/check github.com/jleben/slack-chat-resource/read/check
RUN go build -o /assets/in github.com/jleben/slack-chat-resource/read/in
RUN go build -o /assets/out github.com/jleben/slack-chat-resource/read/out

FROM alpine
RUN apk update
RUN apk add ca-certificates
COPY --from=build-read /assets /opt/resource
