.PHONY: up-go
up-go:
	docker-compose -f docker-compose-dev.yml up --build -- backend-go mysql-backend jiaapi-mock

.PHONY: up-go-detach
up-go-detach:
	docker-compose -f docker-compose-dev.yml up -d --build -- backend-go mysql-backend jiaapi-mock

.PHONY: down-go
down-go:
	docker-compose -f docker-compose-dev.yml down

.PHONY: up-bench
up-bench:
	target=go docker compose -f docker-compose-bench.yml up --build

.PHONY: up-bench-detach
up-bench-detach:
	target=go docker compose -f docker-compose-bench.yml up -d --build

.PHONY: run-bench
run-bench:
	target=go docker compose -f docker-compose-bench.yml exec apitest go run ./main.go \
		--target nginx-backend --tls \
		--tls-cert=../extra/certificates/jiaapi.crt --tls-key=../extra/certificates/jiaapi.key

.PHONY: run-bench-race
run-bench-race:
	target=go docker compose -f docker-compose-bench.yml exec apitest go run -race ./main.go \
		--target nginx-backend --tls \
		--tls-cert=../extra/certificates/jiaapi.crt --tls-key=../extra/certificates/jiaapi.key

.PHONY: run-bench-noload
run-bench-noload:
	target=go docker compose -f docker-compose-bench.yml exec apitest go run ./main.go \
		--target nginx-backend --tls \
		--tls-cert=../extra/certificates/jiaapi.crt --tls-key=../extra/certificates/jiaapi.key --no-load

.PHONY: down-bench
down-bench:
	target=go docker compose -f docker-compose-bench.yml down

