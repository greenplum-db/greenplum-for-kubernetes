package kustomize

import (
	"io"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
)

// Build will build the kustomization directory and return the yaml
func Build(kustomizePath string, out io.Writer) error {
	// From sigs.k8s.io/kustomize/kustomize/v3/internal/commands/build/build.go#RunBuild()
	fSys := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	m, err := k.Run(fSys, kustomizePath)
	if err != nil {
		return err
	}
	// From build.emitResources()
	res, err := m.AsYaml()
	if err != nil {
		return err
	}
	_, err = out.Write(res)
	return err
}

// ExtractCRD will take kustomize output and return the CRD object for greenplumPXF
func ExtractCRD(kustomizeOutput io.Reader, crdName string) (*apiextensionsv1.CustomResourceDefinition, error) {
	y := yaml.NewYAMLOrJSONDecoder(kustomizeOutput, 1024)
	for {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		err := y.Decode(crd)
		if err != nil {
			return nil, err
		}
		if crd.Name == crdName {
			return crd, nil
		}
	}
}
