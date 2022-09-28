package startContainerUtils

import (
	"github.com/pivotal/greenplum-for-kubernetes/pkg/multidaemon"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type MultidaemonStarter struct {
	Daemons []multidaemon.DaemonFunc
}

var _ starter.Starter = &MultidaemonStarter{}

// This is not tested - same rationale as not unit testing in greenplumOperator/main.go
func (m *MultidaemonStarter) Run() error {
	if errs := multidaemon.InitializeDaemons(ctrl.SetupSignalHandler(), m.Daemons...); len(errs) > 0 {
		return k8serrors.NewAggregate(errs)
	}
	return nil
}
