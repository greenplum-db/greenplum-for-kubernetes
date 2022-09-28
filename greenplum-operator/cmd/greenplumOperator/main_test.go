package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetInstanceImageFromEnv", func() {
	var env map[string]string
	getenv := func(key string) string {
		return env[key]
	}

	When("getenv returns values for GREENPLUM_IMAGE_REPO and GREENPLUM_IMAGE_TAG", func() {
		BeforeEach(func() {
			env = map[string]string{
				"GREENPLUM_IMAGE_REPO": "org/image_repo",
				"GREENPLUM_IMAGE_TAG":  "v100",
			}
		})
		It("returns the instanceImage name", func() {
			Expect(GetInstanceImageFromEnv(getenv)).To(Equal("org/image_repo:v100"))
		})
	})

	When("getenv returns only for GREENPLUM_IMAGE_REPO", func() {
		BeforeEach(func() {
			env = map[string]string{
				"GREENPLUM_IMAGE_REPO": "org/image_repo",
				"GREENPLUM_IMAGE_TAG":  "",
			}
		})
		It("returns an error", func() {
			instanceImage, err := GetInstanceImageFromEnv(getenv)
			Expect(instanceImage).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("GREENPLUM_IMAGE_TAG cannot be empty"))
		})
	})

	When("getenv returns only for GREENPLUM_IMAGE_TAG", func() {
		BeforeEach(func() {
			env = map[string]string{
				"GREENPLUM_IMAGE_REPO": "",
				"GREENPLUM_IMAGE_TAG":  "v100",
			}
		})
		It("returns an error", func() {
			instanceImage, err := GetInstanceImageFromEnv(getenv)
			Expect(instanceImage).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("GREENPLUM_IMAGE_REPO cannot be empty"))
		})
	})
})

var _ = Describe("GetOperatorImageFromEnv", func() {
	var env map[string]string
	getenv := func(key string) string {
		return env[key]
	}

	When("getenv returns values for OPERATOR_IMAGE_REPO and OPERATOR_IMAGE_TAG", func() {
		BeforeEach(func() {
			env = map[string]string{
				"OPERATOR_IMAGE_REPO": "org/image_repo",
				"OPERATOR_IMAGE_TAG":  "v100",
			}
		})
		It("returns the instanceImage name", func() {
			Expect(GetOperatorImageFromEnv(getenv)).To(Equal("org/image_repo:v100"))
		})
	})

	When("getenv returns only for OPERATOR_IMAGE_REPO", func() {
		BeforeEach(func() {
			env = map[string]string{
				"OPERATOR_IMAGE_REPO": "org/image_repo",
				"OPERATOR_IMAGE_TAG":  "",
			}
		})
		It("returns an error", func() {
			instanceImage, err := GetOperatorImageFromEnv(getenv)
			Expect(instanceImage).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("OPERATOR_IMAGE_TAG cannot be empty"))
		})
	})

	When("getenv returns only for OPERATOR_IMAGE_TAG", func() {
		BeforeEach(func() {
			env = map[string]string{
				"OPERATOR_IMAGE_REPO": "",
				"OPERATOR_IMAGE_TAG":  "v100",
			}
		})
		It("returns an error", func() {
			instanceImage, err := GetOperatorImageFromEnv(getenv)
			Expect(instanceImage).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("OPERATOR_IMAGE_REPO cannot be empty"))
		})
	})
})
