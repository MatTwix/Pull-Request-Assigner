include .env
export

compose-up: ### Run docker-compose
	docker-compose up --build -d
.PHONY: compose-up

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

compose-down-v: ### Down docker-compose and remove volumes
	docker-compose down -v --remove-orphans
.PHONY: compose-down

docker-rm-volume: ### remove docker volume
	docker volume rm postgres-data
.PHONY: docker-rm-volume

linter: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations 'pr_management'
.PHONY: migrate-create

test: ### run k6 tests
	$(MAKE) compose-up
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/smoke.js 
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/team_tests.js 
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/user_tests.js 
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/pr_tests.js
	$(MAKE) compose-down
.PHONY: test

test-no-compose: ### run k6 tests without compose-up and compose-down
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/smoke.js 
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/team_tests.js 
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/user_tests.js 
	APP_PORT=$(APP_PORT) ADMIN_API_KEY=$(ADMIN_API_KEY) k6 run tests/k6/pr_tests.js
.PHONY: test-no-compose

swag: ### generate swagger docs
	swag init -g internal/app/app.go --parseInternal --parseDependency
