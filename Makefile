user=xnok

all: read-resource post-resource search-resource

read-resource:
	docker build -t ${user}/slack-read-resource -f read/Dockerfile .

post-resource:
	docker build -t ${user}/slack-post-resource -f post/Dockerfile .

search-resource:
	docker build -t ${user}/slack-search-resource -f search/Dockerfile .

test-tt:
	bash test/test-${task}.sh ${type} ${user} ./test/${type}/${task}.json
