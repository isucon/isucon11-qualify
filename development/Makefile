.PHONY: up-go
up-go:
	docker-compose -f docker-compose-dev.yml up --build -- backend-go mysql-backend jiaapi-mock

.PHONY: up-php
up-php:
	docker-compose -f docker-compose-dev.yml up -d --build -- backend-php mysql-backend jiaapi-mock

.PHONY: down-dev
down-dev:
	docker-compose -f docker-compose-dev.yml down

.PHONY: test-go
test-go:
	target=go make test

.PHONY: test
test:
	docker-compose -f docker-compose-ci.yml -p isucon11q-ci up -d --build backend mysql-backend
	docker-compose -f docker-compose-ci.yml -p isucon11q-ci up --build apitest
	docker-compose -f docker-compose-ci.yml -p isucon11q-ci stop

.PHONY: up-bench
up-bench:
	target=go docker compose -f docker-compose-bench.yml up --build

.PHONY: run-bench
run-bench:
	target=go docker compose -f docker-compose-bench.yml exec apitest go run ./main.go

.PHONY: run-bench-race
run-bench-race:
	target=go docker compose -f docker-compose-bench.yml exec apitest go run -race ./main.go

.PHONY: down-bench
down-bench:
	target=go docker compose -f docker-compose-bench.yml down

.PHONY: run-bench-noload
run-bench-noload:
	target=go docker compose -f docker-compose-bench.yml exec apitest go run ./main.go --no-load=true

.PHONY: update-assets
update-assets:
	rm -rf ../webapp/public
	cd ../webapp/frontend && npm run build
	cd ../bench && go generate ./gen/assets.go