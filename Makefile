.PHONY: bin
bin:
	@mkdir -p bin


.PHONY: build-cq
build-cq: bin
	@go build -o bin/cq-source-statedeadlock ./cq-source-statedeadlock/

.PHONY: up
up:
	@docker-compose up -d

.PHONY: down
down:
	@docker-compose down

