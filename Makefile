SHELL:=/bin/bash
STATICCHECK=$(shell which statictest)

test:
	go test -cover ./...

vet:
	go vet ./...
	go vet -vettool=$(STATICCHECK) ./...

buildserver:
	go build  -C cmd/server .

buildagent:
	go build -C cmd/agent .

runserver:
	go run ./cmd/server

runagent:
	go run ./cmd/agent

autotest: autotest1 autotest2 autotest3 autotest4 autotest5

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

autotest5:
	SERVER_PORT=8090 ADDRESS="localhost:8090" TEMP_FILE="/tmp/123"  metricstest -test.v -test.run=^TestIteration5$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8090 \
	-source-path=.