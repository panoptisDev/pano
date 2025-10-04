.PHONY: all
all: panod panotool

GOPROXY ?= "https://proxy.golang.org,direct"
.PHONY: panod panotool
panod:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w -X github.com/panoptisDev/pano/version.gitCommit=$${GIT_COMMIT} \
	                    -X github.com/panoptisDev/pano/version.gitDate=$${GIT_DATE}" \
	    -o build/panod \
	    ./cmd/panod && \
	    ./build/panod version

panotool:
	GIT_COMMIT=`git rev-list -1 HEAD 2>/dev/null || echo ""` && \
	GIT_DATE=`git log -1 --date=short --pretty=format:%ct 2>/dev/null || echo ""` && \
	GOPROXY=$(GOPROXY) \
	go build \
	    -ldflags "-s -w -X github.com/panoptisDev/pano/version.gitCommit=$${GIT_COMMIT} \
	                    -X github.com/panoptisDev/pano/version.gitDate=$${GIT_DATE}" \
	    -o build/panotool \
	    ./cmd/panotool && \
	    ./build/panotool --version

TAG ?= "latest"
.PHONY: pano-image
pano-image:
	docker build \
		--network=host \
		-f ./Dockerfile -t "pano:$(TAG)" .

.PHONY: test
test:
	go test --timeout 30m ./...

.PHONY: coverage
coverage:
	@mkdir -p build ;\
	go test -coverpkg=./... --timeout=30m -coverprofile=build/coverage.cov ./... && \
	go tool cover -html build/coverage.cov -o build/coverage.html &&\
	echo "Coverage report generated in build/coverage.html"

.PHONY: clean
clean:
	rm -fr ./build/*

# Linting

.PHONY: lint
lint: 
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
	@golangci-lint run ./...

.PHONY: license-check
license-check:
	go run ./scripts/license/add_license_header.go --check -dir ./

.PHONY: license-add
license-add:
	go run ./scripts/license/add_license_header.go -dir ./