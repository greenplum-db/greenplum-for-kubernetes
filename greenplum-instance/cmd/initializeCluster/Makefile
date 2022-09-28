INITIALIZE_CLUSTER_VERSION_STR="-X github.com/pivotal/greenplum-for-kubernetes/initializeCluster.version=$(GIT_VERSION)"

.PHONY: build
build: *.go
	go build -tags 'initializeCluster' $(GOFLAGS) -ldflags $(INITIALIZE_CLUSTER_VERSION_STR)

.PHONY: clean
clean:
	rm -f initializeCluster
