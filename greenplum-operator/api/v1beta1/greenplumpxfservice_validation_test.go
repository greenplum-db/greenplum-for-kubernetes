package v1beta1_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1beta1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1beta1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/kustomize"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

var _ = Describe("GreenplumPXFService validation", func() {
	var (
		greenplumPXFCrd *apiextensionsv1.CustomResourceDefinition
		apiCrd          *apiextensions.CustomResourceDefinition
		greenplumPXF    *greenplumv1beta1.GreenplumPXFService
		validator       *validate.SchemaValidator
	)

	BeforeEach(func() {
		b := bytes.NewBuffer(nil)
		Expect(kustomize.Build("../../config/crd", b)).To(Succeed())

		var err error
		greenplumPXFCrd, err = kustomize.ExtractCRD(b, "greenplumpxfservices.greenplum.pivotal.io")
		Expect(err).NotTo(HaveOccurred())

		apiCrd = &apiextensions.CustomResourceDefinition{}
		Expect(scheme.Scheme.Convert(greenplumPXFCrd, apiCrd, nil)).To(Succeed())

		// Convert CRD schema to openapi schema
		validator, _, err = validation.NewSchemaValidator(apiCrd.Spec.Validation)
		Expect(err).NotTo(HaveOccurred())

		greenplumPXF = &greenplumv1beta1.GreenplumPXFService{
			Spec: greenplumv1beta1.GreenplumPXFServiceSpec{
				Replicas: 2,
			},
		}
	})

	It("validates Greenplum PXF", func() {
		Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
	})

	It("sets default values", func() {
		spec := apiCrd.Spec.Validation.OpenAPIV3Schema.Properties["spec"]
		Expect(spec.Properties["replicas"].Default).To(Equal(heapvalue.NewJSONNumber(2)))
	})

	It("adds itself to the `all` category", func() {
		Expect(apiCrd.Spec.Names.Categories).To(ContainElement("all"))
	})

	Context("GreenplumPXF properties", func() {
		It("allow replicas to not be specified", func() {
			greenplumPXF.Spec.Replicas = 0
			Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
		})
		It("allows 1-1000 replicas", func() {
			for i := int32(1); i <= 1000; i *= 2 {
				greenplumPXF.Spec.Replicas = i
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
			}
			greenplumPXF.Spec.Replicas = 1000
			Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
		})
		It("does not allow replicas > 1000", func() {
			greenplumPXF.Spec.Replicas = 1001
			Expect(validator.Validate(greenplumPXF).IsValid()).To(BeFalse())
			Expect(validator.Validate(greenplumPXF).AsError()).To(
				MatchError("validation failure list:\nspec.replicas in body should be less than or equal to 1000"),
				"%#v", validator.Validate(greenplumPXF).AsError().Error())
		})

		When("pxfConf is populated", func() {
			It("does not allow Bucket to be empty", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "",
						EndPoint: "not-empty",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumPXF).AsError()).To(
					MatchError("validation failure list:\nspec.pxfConf.s3Source.bucket in body should be at least 1 chars long"),
					"%#v", validator.Validate(greenplumPXF).AsError().Error())
			})
			It("does not allow Secret to be empty", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "",
						Bucket:   "not-empty",
						EndPoint: "not-empty",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumPXF).AsError()).To(
					MatchError("validation failure list:\nspec.pxfConf.s3Source.secret in body should be at least 1 chars long"),
					"%#v", validator.Validate(greenplumPXF).AsError().Error())
			})
			It("does not allow EndPoint to be empty", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "not-empty",
						EndPoint: "",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumPXF).AsError()).To(
					MatchError("validation failure list:\nspec.pxfConf.s3Source.endpoint in body should be at least 1 chars long"),
					"%#v", validator.Validate(greenplumPXF).AsError().Error())
			})
			It("allows Protocol to be empty", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "not-empty",
						EndPoint: "not-empty",
						Protocol: "",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
			})
			It("allows Protocol to be http", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "not-empty",
						EndPoint: "not-empty",
						Protocol: "http",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
			})
			It("allows Protocol to be https", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "not-empty",
						EndPoint: "not-empty",
						Protocol: "https",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
			})
			It("disllows Protocol to be other values", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "not-empty",
						EndPoint: "not-empty",
						Protocol: "ftp",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeFalse())
				Expect(validator.Validate(greenplumPXF).AsError()).To(
					MatchError("validation failure list:\nspec.pxfConf.s3Source.protocol in body should be one of [http https]"),
					"%#v", validator.Validate(greenplumPXF).AsError().Error())
			})
			It("allows Folder to be empty", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "not-empty",
						EndPoint: "not-empty",
						Folder:   "",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
			})
			It("validates when all properties have non-empty values", func() {
				greenplumPXF.Spec.PXFConf = &greenplumv1beta1.GreenplumPXFConf{
					S3Source: greenplumv1beta1.S3Source{
						Secret:   "not-empty",
						Bucket:   "not-empty",
						EndPoint: "not-empty",
						Protocol: "http",
						Folder:   "not-empty",
					},
				}
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
			})
		})
		When("pxfConf is omitted", func() {
			BeforeEach(func() {
				greenplumPXF.Spec.PXFConf = nil
			})
			It("passes validation", func() {
				Expect(validator.Validate(greenplumPXF).IsValid()).To(BeTrue())
			})
		})
	})
})
