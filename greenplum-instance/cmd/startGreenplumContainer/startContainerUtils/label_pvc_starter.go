package startContainerUtils

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/cppforlife/go-semi-semantic/version"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/hostpod"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LabelPvcStarter struct {
	*starter.App
	Hostname  func() (string, error)
	NewClient func() (client.Client, error)
}

func (s *LabelPvcStarter) Run() error {
	ctx := context.Background()

	c, err := s.NewClient()
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	namespace, err := hostpod.GetCurrentNamespace(s.Fs)
	if err != nil {
		return err
	}
	thisPod, err := hostpod.GetThisPod(ctx, c, namespace, s.Hostname)
	if err != nil {
		return err
	}

	var pvcName string
	for _, volume := range thisPod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			if pvcName != "" {
				return errors.New("found more pvc volumes than expected")
			}
			pvcName = volume.PersistentVolumeClaim.ClaimName
		}
	}

	gpMaj, err := s.GetGreenplumMajorVersion()
	if err != nil {
		return err
	}

	var pvc corev1.PersistentVolumeClaim
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: pvcName}, &pvc); err != nil {
		return err
	}

	if pvc.Labels == nil {
		pvc.Labels = make(map[string]string)
	}
	if _, ok := pvc.Labels["greenplum-major-version"]; !ok {
		origPvc := pvc.DeepCopy()
		pvc.Labels["greenplum-major-version"] = gpMaj
		if err := c.Patch(ctx, &pvc, client.MergeFrom(origPvc)); err != nil {
			return fmt.Errorf("patching pvc label: %w", err)
		}
	}

	if pvc.Labels["greenplum-major-version"] != gpMaj {
		return fmt.Errorf("GPDB version on PVC does not match pod version. PVC greenplum-major-version=%s; Pod version: %s",
			pvc.Labels["greenplum-major-version"], gpMaj)
	}

	return nil
}

var versionRe = regexp.MustCompile(`postgres \(Greenplum Database\) ([0-9A-Za-z_.]+(?:-[0-9A-Za-z_\-.]+)?(?:\+[0-9A-Za-z_\-.]+)?) build commit:[0-9a-f]+`)

func (s *LabelPvcStarter) GetGreenplumMajorVersion() (string, error) {
	cmd := cluster.NewGreenplumCommand(s.Command).Command("/usr/local/greenplum-db/bin/postgres", "--gp-version")
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("trying to get greenplum version: %w; stderr: %s", ee, strings.TrimSpace(string(ee.Stderr)))
		}
		// NB: non-ExitError case is untested
		return "", fmt.Errorf("trying to get greenplum version: %w", err)
	}
	var gpVer string

	m := versionRe.FindStringSubmatch(string(out))
	parseErr := fmt.Errorf("couldn't parse greenplum version in: %s", strings.TrimSpace(string(out)))
	if len(m) < 2 {
		return "", parseErr
	}
	gpVer = m[1]

	semver, err := version.NewVersionFromString(gpVer)
	if err != nil {
		return "", parseErr
	}
	if len(semver.Release.Components) == 0 {
		return "", parseErr
	}
	gpMaj := semver.Release.Components[0].AsString()

	return gpMaj, nil
}
