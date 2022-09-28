package testing

import (
	"github.com/pivotal/greenplum-for-kubernetes/pkg/instanceconfig"
)

type MockReader struct {
	NamespaceName    string
	NamespaceNameErr error

	GreenplumClusterName    string
	GreenplumClusterNameErr error

	SegmentCount    int
	SegmentCountErr error

	Mirrors    bool
	MirrorsErr error

	PXFServiceName    string
	PXFServiceNameErr error

	Standby    bool
	StandbyErr error

	ConfigMapValuesErr error
}

var _ instanceconfig.Reader = &MockReader{}

func (cr *MockReader) GetNamespace() (string, error) {
	return cr.NamespaceName, cr.NamespaceNameErr
}

func (cr *MockReader) GetGreenplumClusterName() (string, error) {
	return cr.GreenplumClusterName, cr.GreenplumClusterNameErr
}

func (cr *MockReader) GetSegmentCount() (count int, err error) {
	return cr.SegmentCount, cr.SegmentCountErr
}

func (cr *MockReader) GetMirrors() (bool, error) {
	return cr.Mirrors, cr.MirrorsErr
}

func (cr *MockReader) GetPXFServiceName() (string, error) {
	return cr.PXFServiceName, cr.PXFServiceNameErr
}

func (cr *MockReader) GetStandby() (bool, error) {
	return cr.Standby, cr.StandbyErr
}

func (cr *MockReader) GetConfigValues() (instanceconfig.ConfigValues, error) {
	return instanceconfig.ConfigValues{
		Namespace:            cr.NamespaceName,
		GreenplumClusterName: cr.GreenplumClusterName,
		SegmentCount:         cr.SegmentCount,
		Mirrors:              cr.Mirrors,
		Standby:              cr.Standby,
		PXFServiceName:       cr.PXFServiceName,
	}, cr.ConfigMapValuesErr
}
