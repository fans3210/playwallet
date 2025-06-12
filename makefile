.PHONY: lint
lint:
	golangci-lint run -c golangci.yml ./... -v

.PHONY: run
run:
	docker-compose --project-name playwallet -f docker-compose.yaml up -d 

.PHONY: dev
dev:
	docker-compose --project-name playwallet -f docker-compose.yaml -f docker-compose.dev.yaml up -d --build

.PHONY: down
down:
	docker-compose --project-name playwallet -f docker-compose.yaml down

.PHONY: build
build:
	docker buildx build --platform linux/amd64 -t github.com/fans3210/playwallet -f Dockerfile .


# generate test data 
.PHONY: testdata
testdata: 
	docker exec -it walletpg psql -h postgres -U admin -d playwallet -f ./testdata.sql

.PHONY: e2etest
e2etest:
	go test ./tests -count=1

.PHONY: e2etestv
e2etestv:
	go test ./tests -v -count=1
