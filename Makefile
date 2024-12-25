include .envrc 
MIGRATIONS_PATH = /db/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY: migrate-create
migration:
	@docker run -it --rm -v ./internal/db:/db --network host migrate/migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@docker run -it --rm --network host --volume ./internal/db:/db migrate/migrate -path=$(MIGRATIONS_PATH) -database $(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@docker run -it --rm --network host --volume ./internal/db:/db migrate/migrate -path=$(MIGRATIONS_PATH) -database $(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run ./internal/db/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./app/main.go -d cmd,internal && swag fmt