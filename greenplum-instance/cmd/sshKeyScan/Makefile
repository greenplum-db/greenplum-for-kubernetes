SSH_KEYSSCAN_VERSION_STR="-X github.com/pivotal/greenplum-for-kubernetes/sshKeyScan.version=$(GIT_VERSION)"

.PHONY: build
build: *.go
	go build -tags 'sshKeyScan' $(GOFLAGS) -ldflags $(SSH_KEYSSCAN_VERSION_STR)

.PHONY: clean
clean:
	rm -f sshKeyScan
