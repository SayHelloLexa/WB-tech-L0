PROJECT_NAME=wb-service

BIN_APP=bin/app
BIN_PRODUCER=bin/producer
BIN_CONSUMER=bin/consumer

SRC_APP=cmd/app/main.go
SRC_PRODUCER=cmd/producer/main.go
SRC_CONSUMER=cmd/consumer/main.go

.PHONY: start-compose build run-all stop clean

setup: start-compose build run-all

start-compose:
	docker-compose -p ${PROJECT_NAME} up

build:
	go build -o ${BIN_APP} ${SRC_APP}
	go build -o ${BIN_PRODUCER} ${SRC_PRODUCER}
	go build -o ${BIN_CONSUMER} ${SRC_CONSUMER}

run-all: build
	./${BIN_APP} & echo $$! > ${BIN_APP}.pid
	./${BIN_PRODUCER} & echo $$! > ${BIN_PRODUCER}.pid
	./${BIN_CONSUMER} & echo $$! > ${BIN_CONSUMER}.pid

stop:
	@if [ -f ${BIN_APP}.pid ]; then kill `cat ${BIN_APP}.pid` || true; rm ${BIN_APP}.pid; fi
	@if [ -f ${BIN_PRODUCER}.pid ]; then kill `cat ${BIN_PRODUCER}.pid` || true; rm ${BIN_PRODUCER}.pid; fi
	@if [ -f ${BIN_CONSUMER}.pid ]; then kill `cat ${BIN_CONSUMER}.pid` || true; rm ${BIN_CONSUMER}.pid; fi
	docker-compose -p ${PROJECT_NAME} down


clean: stop
	rm -r bin

help:
	@echo "===============Available targets==============="
	@echo ""
	@echo "start compose	-- Run docker-compose file"
	@echo "build		-- Compile go-files into binaries"
	@echo "run-all		-- Run all service components"
	@echo "stop		-- Stop service"
	@echo "clean		-- Clean binaries"
	@echo ""
	@echo "==============================================="