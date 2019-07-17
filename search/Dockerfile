FROM golang:alpine as build-env
RUN apk update
RUN apk add git
RUN go get -u github.com/golang/dep/cmd/dep
COPY ./utils /go/src/github.com/jleben/slack-chat-resource/utils
COPY ./slack /go/src/github.com/nlopes/slack
# COPY Gopkg.toml /go/src/github.com/jleben/slack-chat-resource
# RUN ( cd /go/src/github.com/jleben/slack-chat-resource ; dep ensure )
RUN go build -o /assets/utils github.com/jleben/slack-chat-resource/utils

FROM build-env as build-search
COPY ./search/check /go/src/github.com/jleben/slack-chat-resource/search/check/
COPY ./search/in /go/src/github.com/jleben/slack-chat-resource/search/in/
COPY ./search/out /go/src/github.com/jleben/slack-chat-resource/search/out/
RUN go build -o /assets/check github.com/jleben/slack-chat-resource/search/check
RUN go build -o /assets/in github.com/jleben/slack-chat-resource/search/in
RUN go build -o /assets/out github.com/jleben/slack-chat-resource/search/out

FROM alpine
RUN apk update
RUN apk add ca-certificates
COPY --from=build-search /assets /opt/resource
