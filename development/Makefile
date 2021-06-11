.PHONY: up
up:
	docker-compose -f docker-compose.yml up -d --build mysql-backend backend-go

.PHONY: apitest
apitest:
	target=backend-go docker-compose -f docker-compose.yml up --build apitest

.PHONY: down
down:
	docker-compose -f docker-compose.yml down
