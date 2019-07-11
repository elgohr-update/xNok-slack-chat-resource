#! /bin/bash
type=$1
user=$2
request=$3

if [[ -z $type || -z $request ]]; then
    echo "Required arguments: <resource type> <request file>"
    exit 1
fi

cat "$request" | docker run --rm -i $user/slack-$type-resource /opt/resource/check
