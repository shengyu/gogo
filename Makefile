APP := gogo
DEV_IMAGE := $(APP):dev
KIND_CLUSTER ?= kind

.PHONY: run test mock mock-transient mock-test fmt tidy build-image render-dev render-prod deploy-dev deploy-prod kind-load-dev clean-dev

run:
	go run ./cmd/server

test:
	go test ./...

mock:
	cd mock && MOCK_SCENARIO=normal go run .

mock-transient:
	cd mock && MOCK_SCENARIO=transient go run .

mock-test:
	cd mock && go test ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

build-image:
	docker build -t $(DEV_IMAGE) --build-arg VERSION=dev --build-arg COMMIT=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) .

render-dev:
	kubectl kustomize deploy/overlays/dev

render-prod:
	kubectl kustomize deploy/overlays/prod

kind-load-dev: build-image
	kind load docker-image $(DEV_IMAGE) --name $(KIND_CLUSTER)

deploy-dev:
	kubectl apply -k deploy/overlays/dev
	kubectl rollout status deployment/gogo -n gogo-dev

deploy-prod:
	kubectl apply -k deploy/overlays/prod

clean-dev:
	kubectl delete -k deploy/overlays/dev --ignore-not-found
