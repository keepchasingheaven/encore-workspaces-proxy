OS = $(shell uname | tr A-Z a-z)
SSH_HOST_KEY_PATH = $$(pwd)/localdev-ssh-host-key

# versões de dependência
GOTESTSUM_VERSION = 0.6.0

.PHONY: build test run docker run-backends clean all lint fmt vet tidy coverage

all: clean build run

clean:
	@rm -f ./proxy

lint: bin/golangci-lint tidy fmt
	@./bin/golangci-lint run

vet:
	@go vet ./...

fmt:
	@go fmt ./...

build:
	@mkdir -p bin
	@go build -o proxy

tidy:
	@go mod tidy

test: bin/gotestsum-${GOTESTSUM_VERSION}
	@./bin/gotestsum-${GOTESTSUM_VERSION} --no-summary=skipped --junitfile ./coverage.xml --format short-verbose -- -coverprofile=./coverage.txt -covermode=atomic ./...

run: build
	[ -f ${SSH_HOST_KEY_PATH} ] && echo "SSH host key exists" || ssh-keygen -f ${SSH_HOST_KEY_PATH} -N '' -t rsa
	SSH_HOST_KEY="$$(cat ${SSH_HOST_KEY_PATH})" ./proxy --http-enabled=true --ssh-enabled=true --auth-host="http://gdk.test:3000" --auth-client-id="${AUTH_CLIENT_ID}" --auth-client-secret="${AUTH_CLIENT_SECRET}" --auth-redirect-uri="http://127.0.0.1:9876/auth/callback" --auth-signing-key="example-auth-signing-key" --kubeconfig="$$HOME/.kube/config"

coverage:
	go tool cover -func

bin/gotestsum-${GOTESTSUM_VERSION}:
	@mkdir -p bin
	@curl -L https://github.com/gotestyourself/gotestsum/releases/download/v${GOTESTSUM_VERSION}/gotestsum_${GOTESTSUM_VERSION}_${OS}_amd64.tar.gz | tar -zOxf - gotestsum > ./bin/gotestsum-${GOTESTSUM_VERSION} && chmod +x ./bin/gotestsum-${GOTESTSUM_VERSION}

bin/golangci-lint:
	@mkdir -p bin
	@curl -sSfL https://golangci-lint.run/install.sh | sh -s

run-backends:
	docker rm -vf nginx && \
	docker run -d --name nginx -p 8090:80 nginx && \
	docker rm -vf ttyd && \
	docker run -d --name ttyd -p 8091:7681 tsl0922/ttyd && \
	docker rm -vf vscode && \
	docker run -d --name vscode -p 8092:3000 gitpod/openvscode-server

docker-build-and-publish:
	@./scripts/build_and_publish_image.sh "${USER}" "${TOKEN}" "${IMAGE_NAME}" "${IMAGE_VERSION}"

helm-build-and-publish:
	@./scripts/build_and_publish_chart.sh "${USER}" "${TOKEN}" "${CHART_NAME}" "${CHART_VERSION}"

deploy-local-changes:
	@./scripts/deploy_local_changes.sh

revert-local-to-version:
	@helm upgrade encore-workspaces-proxy encore-workspaces-proxy/encore-workspaces-proxy --version=$(VERSION) -f ./localdev/config.yaml -n encore-workspaces

clean-dev-images:
	$(eval DEV_IMAGES=$(shell nerdctl images --namespace k8s.io --format "{{.ID}} {{.Repository}}:{{.Tag}}" | grep 'registry.github.com/keepchasingheaven/encore-workspaces-proxy:dev' | awk '{print $$1}'))
	@nerdctl rmi -f --namespace k8s.io $(DEV_IMAGES)
