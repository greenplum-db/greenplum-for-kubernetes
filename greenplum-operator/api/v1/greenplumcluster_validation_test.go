package v1_test

import (
	"bytes"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/kustomize"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

var _ = Describe("GreenplumCluster validation", func() {
	var (
		greenplumClusterCRD *apiextensionsv1.CustomResourceDefinition
		apiCrd              *apiextensions.CustomResourceDefinition
		greenplumCluster    *greenplumv1.GreenplumCluster
		validator           *validate.SchemaValidator
	)

	var yesAndNoes = []string{"yes", "Yes", "YES", "no", "No", "NO"}

	BeforeEach(func() {
		b := bytes.NewBuffer(nil)
		Expect(kustomize.Build("../../config/crd", b)).To(Succeed())

		var err error
		greenplumClusterCRD, err = kustomize.ExtractCRD(b, "greenplumclusters.greenplum.pivotal.io")
		Expect(err).NotTo(HaveOccurred())

		apiCrd = &apiextensions.CustomResourceDefinition{}
		Expect(scheme.Scheme.Convert(greenplumClusterCRD, apiCrd, nil)).To(Succeed())

		// Convert CRD schema to openapi schema
		validator, _, err = validation.NewSchemaValidator(apiCrd.Spec.Validation)
		Expect(err).NotTo(HaveOccurred())

		fakeNameSpace := metav1.ObjectMeta{
			Namespace:       "test",
			Name:            "my-greenplum",
			ResourceVersion: time.Now().String(),
		}
		fakeGreenplumMasterAndStandbySpec := greenplumv1.GreenplumMasterAndStandbySpec{
			GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
				Memory:           resource.MustParse("800Mi"),
				CPU:              resource.MustParse("0.5"),
				StorageClassName: "default",
				Storage:          resource.MustParse("1M"),
				WorkerSelector:   map[string]string{},
			},
		}
		fakeGreenplumSegmentSpec := greenplumv1.GreenplumSegmentsSpec{
			GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
				Memory:           resource.MustParse("900Mi"),
				CPU:              resource.MustParse("0.5"),
				StorageClassName: "default",
				Storage:          resource.MustParse("1M"),
				WorkerSelector:   map[string]string{},
			},
			PrimarySegmentCount: 1,
		}
		greenplumCluster = &greenplumv1.GreenplumCluster{
			ObjectMeta: fakeNameSpace,
			TypeMeta: metav1.TypeMeta{
				Kind:       "GreenplumCluster",
				APIVersion: "v1",
			},
			Spec: greenplumv1.GreenplumClusterSpec{
				MasterAndStandby: fakeGreenplumMasterAndStandbySpec,
				Segments:         fakeGreenplumSegmentSpec,
			},
		}
	})

	It("validates GreenplumCluster", func() {
		result := validator.Validate(greenplumCluster)
		Expect(result.IsValid()).To(BeTrue(), fmt.Sprint(result.Errors))
	})

	It("sets default values", func() {
		defaultValueNo := apiextensions.JSON("no")
		spec := apiCrd.Spec.Validation.OpenAPIV3Schema.Properties["spec"]
		masterAndStandbySpec := spec.Properties["masterAndStandby"]
		segmentsSpec := spec.Properties["segments"]
		Expect(masterAndStandbySpec.Properties["antiAffinity"].Default).To(Equal(&defaultValueNo))
		Expect(masterAndStandbySpec.Properties["standby"].Default).To(Equal(&defaultValueNo))
		Expect(segmentsSpec.Properties["antiAffinity"].Default).To(Equal(&defaultValueNo))
		Expect(segmentsSpec.Properties["mirrors"].Default).To(Equal(&defaultValueNo))
	})

	It("sets additionalPrinterColunms", func() {
		expectedAdditionalPrinterColumns := []apiextensions.CustomResourceColumnDefinition{
			{
				Name:        "Status",
				Type:        "string",
				Description: "The greenplum instance status",
				JSONPath:    ".status.phase",
			},
			{
				Name:        "Age",
				Type:        "date",
				Description: "The greenplum instance age",
				JSONPath:    ".metadata.creationTimestamp",
			},
		}
		Expect(apiCrd.Spec.AdditionalPrinterColumns).To(Equal(expectedAdditionalPrinterColumns))
	})

	It("adds itself to the `all` category", func() {
		Expect(apiCrd.Spec.Names.Categories).To(ContainElement("all"))
	})

	Context("GreenplumCluster properties", func() {
		Context("spec.masterAndStandby", func() {
			// storageClassName
			It("Requires at least one character for storageClassName", func() {
				greenplumCluster.Spec.MasterAndStandby.StorageClassName = ""
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.masterAndStandby.storageClassName in body should be at least 1 chars long"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())

				greenplumCluster.Spec.MasterAndStandby.StorageClassName = "default"
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})

			// workerSelector
			It("accepts a not present workerSelector", func() {
				greenplumCluster.Spec.MasterAndStandby.WorkerSelector = map[string]string{}
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})
			It("accepts a valid workerSelector", func() {
				greenplumCluster.Spec.MasterAndStandby.WorkerSelector = map[string]string{"a-valid_Label.99": "foo"}
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})

			// antiAffinity
			It("must be a valid antiAffinity value", func() {
				greenplumCluster.Spec.MasterAndStandby.AntiAffinity = "garbage"
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.masterAndStandby.antiAffinity in body should match '^(?:yes|Yes|YES|no|No|NO|)$'"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())
			})
			It("accepts empty antiAffinity value", func() {
				greenplumCluster.Spec.MasterAndStandby.AntiAffinity = ""
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})
			It("accepts valid yes and no values for antiAffinity", func() {
				for _, value := range yesAndNoes {
					greenplumCluster.Spec.MasterAndStandby.AntiAffinity = value
					Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
				}
			})

			// standby
			It("rejects garbage standby", func() {
				greenplumCluster.Spec.MasterAndStandby.Standby = "garbage value"
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.masterAndStandby.standby in body should match '^(?:yes|Yes|YES|no|No|NO|)$'"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())
			})
			It("accepts empty string standby", func() {
				greenplumCluster.Spec.MasterAndStandby.Standby = ""
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})
			It("accepts all variants of yes and no for standby", func() {
				for _, value := range yesAndNoes {
					greenplumCluster.Spec.MasterAndStandby.Standby = value
					Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
				}
			})
			// required properties
			It("requires storageClassName and storage to be specified", func() {
				required := apiCrd.Spec.Validation.OpenAPIV3Schema.Properties["spec"].Properties["masterAndStandby"].Required
				Expect(required).To(ConsistOf("storageClassName", "storage"))
			})
		})

		Context("spec.segments", func() {
			// primarySegmentCount
			It("does not allow primarySegmentCount < 1", func() {
				greenplumCluster.Spec.Segments.PrimarySegmentCount = 0
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.segments.primarySegmentCount in body should be greater than or equal to 1"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())
			})
			It("allows 1-10000 primarySegmentCount", func() {
				for i := int32(1); i <= 10000; i *= 2 {
					greenplumCluster.Spec.Segments.PrimarySegmentCount = i
					Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
				}
				greenplumCluster.Spec.Segments.PrimarySegmentCount = 10000
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})
			It("does not allow primarySegmentCount > 10000", func() {
				greenplumCluster.Spec.Segments.PrimarySegmentCount = 10001
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.segments.primarySegmentCount in body should be less than or equal to 10000"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())
			})

			// storageClassName
			It("Requires at least one character for storageClassName", func() {
				greenplumCluster.Spec.Segments.StorageClassName = ""
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.segments.storageClassName in body should be at least 1 chars long"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())

				greenplumCluster.Spec.Segments.StorageClassName = "default"
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})

			// workerSelector
			It("accepts a not present workerSelector", func() {
				greenplumCluster.Spec.Segments.WorkerSelector = map[string]string{}
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})
			It("accepts a valid workerSelector", func() {
				greenplumCluster.Spec.Segments.WorkerSelector = map[string]string{"a-valid_Label.99": "foo"}
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})

			// antiAffinity
			It("must be a valid antiAffinity value", func() {
				greenplumCluster.Spec.Segments.AntiAffinity = "garbage"
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.segments.antiAffinity in body should match '^(?:yes|Yes|YES|no|No|NO|)$'"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())
			})
			It("accepts empty antiAffinity value", func() {
				greenplumCluster.Spec.Segments.AntiAffinity = ""
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})
			It("accepts valid yes and no values for antiAffinity", func() {
				for _, value := range yesAndNoes {
					greenplumCluster.Spec.Segments.AntiAffinity = value
					Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
				}
			})

			// mirrors
			It("rejects garbage mirrors", func() {
				greenplumCluster.Spec.Segments.Mirrors = "garbage value"
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumCluster).AsError()).To(
					MatchError("validation failure list:\nspec.segments.mirrors in body should match '^(?:yes|Yes|YES|no|No|NO|)$'"),
					"%#v", validator.Validate(greenplumCluster).AsError().Error())
			})
			It("accepts empty string mirrors", func() {
				greenplumCluster.Spec.Segments.Mirrors = ""
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})
			It("accepts all variants of yes and no for mirrors", func() {
				for _, value := range yesAndNoes {
					greenplumCluster.Spec.Segments.Mirrors = value
					Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
				}
			})

			// required properties
			It("requires primarySegmentCount, storageClassName and storage to be specified", func() {
				required := apiCrd.Spec.Validation.OpenAPIV3Schema.Properties["spec"].Properties["segments"].Required
				Expect(required).To(ConsistOf("primarySegmentCount", "storageClassName", "storage"))
			})
		})

		Describe("PXF validation", func() {
			It("accepts a string value for serviceName", func() {
				greenplumCluster.Spec.PXF.ServiceName = "test-hostname"
				Expect(validator.Validate(greenplumCluster).IsValid()).To(BeTrue())
			})

			It("requires serviceName", func() {
				required := apiCrd.Spec.Validation.OpenAPIV3Schema.Properties["spec"].Properties["pxf"].Required
				Expect(required).To(ConsistOf("serviceName"))
			})
		})

		Context("spec", func() {
			It("requires masterAndStandby and segments to be specified", func() {
				required := apiCrd.Spec.Validation.OpenAPIV3Schema.Properties["spec"].Required
				Expect(required).To(ConsistOf("masterAndStandby", "segments"))
			})
		})

		Context("status", func() {
			It("does not require any properties", func() {
				required := apiCrd.Spec.Validation.OpenAPIV3Schema.Properties["status"].Required
				Expect(required).To(BeEmpty())
			})
		})
	})
})
