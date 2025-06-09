.PHONY: lint
lint:
	golangci-lint run -c golangci.yml ./... -v

.PHONY: dev
dev:
	docker-compose --project-name playwallet -f docker-compose.yaml up -d 

.PHONY: down
down:
	docker-compose --project-name playwallet -f docker-compose.yaml down

.PHONY: build
build:
	go mod tidy
	mkdir -p ./bin && go build -o ./bin ./cmd/playwallet 
