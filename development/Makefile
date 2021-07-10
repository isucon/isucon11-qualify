.PHONY: up-go
up-go:
	docker-compose -f docker-compose-dev.yml up --build -- backend-go mysql-backend jiaapi-mock

.PHONY: down-go
down-go:
	docker-compose -f docker-compose-dev.yml down
