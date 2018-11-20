SHELL:=/bin/bash

image:
	docker build -t cortex/vistar-go-client .
	docker run -it --rm \
		--name vistar-go-client \
		-v "$(CURDIR)":/usr/src/vistar-go-client \
		-w /usr/src/vistar-go-client \
		cortex/vistar-go-client

test:
	richgo test ./... -v -cover -covermode=atomic -parallel 4 -race -timeout 2s

init-dep:
	go mod init github.com/cortexsystems/vistar-go-client

.PHONY: image test init-dep
