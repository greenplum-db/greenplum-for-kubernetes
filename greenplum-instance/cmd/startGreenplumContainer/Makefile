START_GREENPLUM_CONTAINER_VERSION_STR="-X github.com/pivotal/greenplum-for-kubernetes/startGreenplumContainer.version=$(GIT_VERSION)"

.PHONY: build
build: *.go
	go build -tags 'startGreenplumContainer' $(GOFLAGS) -ldflags $(START_GREENPLUM_CONTAINER_VERSION_STR)

.PHONY: clean
clean:
	rm -f startGreenplumContainer
