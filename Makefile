-include .env
.EXPORT_ALL_VARIABLES:

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