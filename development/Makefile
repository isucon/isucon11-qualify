.PHONY: up-go
up-go:
	docker-compose -f docker-compose-dev.yml up --build -- backend-go mysql-backend jiaapi-mock

.PHONY: down-go
down-go:
	docker-compose -f docker-compose-dev.yml down

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