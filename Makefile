SHELL:=/bin/bash
STATICCHECK=$(shell which staticcheck)

.DEFAULT_GOAL := build

test:
	go test -cover ./...

vet:
	go vet ./...
	$(STATICCHECK) ./...

generate:
	go generate ./...

gvt: generate vet test

cover:
	go test -cover ./... -coverprofile cover.out
	go tool cover -html cover.out -o cover.html

build: gvt buildagent buildserver

buildserver:
	go build  -C cmd/server .

buildagent:
	go build -C cmd/agent .

runserver:
	go run ./cmd/server

runagent:
	go run ./cmd/agent

rundb:
	docker compose up -d

racetest:
	go test -v -race ./...

autotest: autotest1 autotest2 autotest3 autotest4 autotest5 autotest6 autotest7 autotest8 autotest9 autotest10 autotest11 autotest12 autotest13 autotest14

autotest1: buildserver
	metricstest -test.v -test.run=^TestIteration1$$ -binary-path=cmd/server/server

autotest2: buildagent
	metricstest -test.v -test.run=^TestIteration2[AB]*$$ \
            -source-path=. \
            -agent-binary-path=cmd/agent/agent

autotest3: buildserver buildagent
	metricstest -test.v -test.run=^TestIteration3[AB]*$$ \
            -source-path=. \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server

autotest4: buildserver buildagent
	SERVER_PORT=8090 ADDRESS="localhost:8090" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration4$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8090 \
	-source-path=.

autotest5: buildserver buildagent
	SERVER_PORT=8090 ADDRESS="localhost:8090" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration5$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8090 \
	-source-path=.

autotest6: buildserver buildagent
	SERVER_PORT=8090 ADDRESS="localhost:8090" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration6$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8090 \
	-source-path=.

autotest7: buildserver buildagent
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration7$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest8: buildserver buildagent
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration8$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest9: buildserver buildagent
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration9$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-file-storage-path=/tmp/123 \
	-server-port=8080 \
	-source-path=. \

autotest10: buildserver buildagent rundb
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration10[AB]$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable' \
	-server-port=8080 \
	-source-path=.

autotest11: buildserver buildagent rundb
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration11$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable' \
	-server-port=8080 \
	-source-path=.

autotest12: buildserver buildagent rundb
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration12$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable' \
	-server-port=8080 \
	-source-path=.

autotest13: buildserver buildagent rundb
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration13$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable' \
	-server-port=8080 \
	-source-path=.

autotest14: buildserver buildagent rundb racetest
	SERVER_PORT=8080 ADDRESS="localhost:8080" TEMP_FILE="/tmp/123" metricstest -test.v -test.run=^TestIteration14$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable' \
	-server-port=8080 \
	-key="supersecret" \
	-source-path=.
