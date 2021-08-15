up-go:
	docker-compose -f docker-compose-dev.yml up --build -- backend-go mysql-backend jiaapi-mock

.PHONY: up-nodejs
up-nodejs:
	docker-compose -f docker-compose-dev.yml up --build -- backend-nodejs mysql-backend jiaapi-mock

.PHONY: up-ruby
up-ruby:
	docker-compose -f docker-compose-dev.yml up --build -- backend-ruby mysql-backend jiaapi-mock

.PHONY: down-dev
down-dev:
	docker-compose -f docker-compose-dev.yml down

.PHONY: test-go
test-go:
	target=go make test

.PHONY: test-nodejs
test-nodejs:
	target=nodejs make test

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

.PHONY: mysql
mysql:
	mysql -h 127.0.0.1 -P $(shell docker-compose -f docker-compose-dev.yml port mysql-backend 3306|cut -d: -f2) -uroot -proot isucondition
