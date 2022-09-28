package integrationutils

import (
	"flag"
	"path"
)

// Version of the operator we use to test upgrades
const OldOperatorVersion = "v2.2.0"

var (
	releaseDir = flag.String("release-dir", "",
		"If set, use artifacts from this release directory. If unset (default), use artifacts from source.")
	oldReleaseDir = flag.String("old-release-dir", "/tmp/greenplum-instance_release/greenplum-for-kubernetes-"+OldOperatorVersion,
		"For upgrade tests, use artifacts from this release directory for an old operator release.")
)

// ReleasePath returns a path relative to source code directory or the release directory.
// If both of those paths are the same, the source code path can be omitted.
func ReleasePath(inRelease string, inSource ...string) string {
	inSourceDir := inRelease
	if len(inSource) > 0 {
		inSourceDir = inSource[0]
	}
	if *releaseDir != "" {
		return path.Join(*releaseDir, inRelease)
	}
	return path.Join("../../..", inSourceDir)
}
