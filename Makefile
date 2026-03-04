-include .env
.EXPORT_ALL_VARIABLES:

.PHONY: migration-create
migration-create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: migration-up
migration-up:
	migrate -path migrations -database "$(MIGRATE_DSN)" up

.PHONY: migration-down
migration-down:
	migrate -path migrations -database "$(MIGRATE_DSN)" down 1

.PHONY: runp
runp: 
	go run ./cmd/producer/main.go

.PHONY: runc
runc: 
	@echo ${KAFKA_TOPICS}
	go run ./cmd/consumer/main.go


.PHONY: proto
proto:
	@buf generate

.PHONY: swagger
swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.8.1 init -g cmd/app.go -o internal/docs      