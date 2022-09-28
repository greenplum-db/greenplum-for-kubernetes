package main

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParsePackageReceipt", func() {
	It("works", func() {
		b := strings.Builder{}
		b.WriteString("Release SHASUM: 6e323391e441d69f5c1b20c56f023e2568c216382a2a3aae7341e5b1527d1c2e\n")
		b.WriteString("\n")
		b.WriteString("ii  dash     0.5.8-2.1ubuntu2   amd64    POSIX-compliant shell\n")
		b.WriteString("# comment?\n")
		b.WriteString("# a longer comment with more words in it?\n")
		b.WriteString("ii  dbus             1.10.6-1ubuntu3.5  amd64    simple interprocess messaging system (daemon and utilities)\n")
		b.WriteString("ii  debconf          1.5.58ubuntu2      all      Debian configuration management system\n")
		b.WriteString("ii  libicu55:amd64   55.1-7ubuntu0.5    amd64    International Components for Unicode\n")
		packages := ParsePackageReceipt(strings.NewReader(b.String()))
		Expect(packages).To(Equal(PackageVersions{
			"dash":     "0.5.8-2.1ubuntu2",
			"dbus":     "1.10.6-1ubuntu3.5",
			"debconf":  "1.5.58ubuntu2",
			"libicu55": "55.1-7ubuntu0.5",
		}))
	})
})

var _ = Describe("ParseUSN", func() {
	var story SecurityStory

	When("story is well-formed", func() {
		BeforeEach(func() {
			b := strings.Builder{}
			b.WriteString("**Product**: Pivotal Greenplum for Kubernetes\n")
			b.WriteString("**Severity**: medium\n")
			b.WriteString("**USN**: https://usn.ubuntu.com/4247-1/\n")
			b.WriteString("**Davos Notice**: https://davos.cfapps.io/notices/USN-4247-1\n")
			b.WriteString("**19.10 Packages:**\n")
			b.WriteString("python-apt 1.9.0ubuntu1.2\n")
			b.WriteString("python3-apt 1.9.0ubuntu1.2\n")
			b.WriteString("**18.04 Packages:**\n")
			b.WriteString("python-apt 1.6.5ubuntu0.1\n")
			b.WriteString("python3-apt 1.6.5ubuntu0.1\n")
			b.WriteString("**16.04 Packages:**\n")
			b.WriteString("python-apt 1.1.0~beta1ubuntu0.16.04.7\n")
			b.WriteString("python3-apt 1.1.0~beta1ubuntu0.16.04.7\n")
			story = SecurityStory{
				Title:       "**[Security Notice]** New USN affecting gp4k: USN-4247-1: python-apt vulnerabilities",
				Description: b.String(),
			}
			Expect(story.ParseUSN()).To(Succeed())
		})
		It("finds the 18.04 packages", func() {
			Expect(story.usn.packages).To(Equal(PackageVersions{
				"python-apt":  "1.6.5ubuntu0.1",
				"python3-apt": "1.6.5ubuntu0.1",
			}))
		})
		It("extracts the USN link", func() {
			Expect(story.usn.link).To(Equal("https://usn.ubuntu.com/4247-1/"))
		})
		It("extracts the USN ID", func() {
			Expect(story.usn.id).To(Equal("USN-4247-1"))
		})
		It("extracts the USN headline", func() {
			Expect(story.usn.headline).To(Equal("python-apt vulnerabilities"))
		})
	})

	When("story does not contain a link", func() {
		BeforeEach(func() {
			b := strings.Builder{}
			b.WriteString("**USN link has a different title now, hee hee**: https://usn.ubuntu.com/4247-1/\n")
			b.WriteString("**18.04 Packages:**\n")
			b.WriteString("python-apt 1.1.0~beta1ubuntu0.16.04.7\n")
			b.WriteString("python3-apt 1.1.0~beta1ubuntu0.16.04.7\n")
			story = SecurityStory{
				ID:          "170000001",
				Title:       "**[Security Notice]** New USN affecting gp4k: USN-4247-1: python-apt vulnerabilities",
				Description: b.String(),
			}

		})
		It("returns an error", func() {
			Expect(story.ParseUSN()).To(MatchError("could not parse USN link from story: " + story.ID))
		})
	})
	When("story title does not contain match the expected format", func() {
		BeforeEach(func() {
			b := strings.Builder{}
			b.WriteString("**USN**: https://usn.ubuntu.com/4247-1/\n")
			b.WriteString("**18.04 Packages:**\n")
			b.WriteString("python-apt 1.1.0~beta1ubuntu0.16.04.7\n")
			b.WriteString("python3-apt 1.1.0~beta1ubuntu0.16.04.7\n")
			story = SecurityStory{
				ID:          "170000001",
				Title:       "CVE-4247-1 python-apt vulnerabilities",
				Description: b.String(),
			}

		})
		It("returns an error", func() {
			Expect(story.ParseUSN()).To(MatchError("could not parse USN title and ID from story: " + story.ID))
		})
	})
})

var _ = Describe("FindPackagesRequiringUpdate", func() {
	var (
		storyPkg PackageVersions
		dpkg     PackageVersions
		story    SecurityStory
	)
	JustBeforeEach(func() {
		story = SecurityStory{
			ID: "12345",
		}
		story.usn.packages = storyPkg
	})
	When("only one story package is in dpkg, and it's up-to-date", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.2",
				"gcc":        "6",
			}
			dpkg = PackageVersions{
				"gcc": "6",
			}
		})
		It("is okay", func() {
			result, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})
	When("dpkg contains no packages that are in the story packages", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"libicu55": "55.1-7ubuntu0.5",
			}
			dpkg = PackageVersions{
				"libicu55:amd64": "55.1-7ubuntu0.5", // We should have removed ":amd64" earlier.
			}
		})
		It("returns an error", func() {
			_, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).To(HaveOccurred())
			// This is a sanity check that we're processing the dpkg list correctly to avoid false positives.
			Expect(err.Error()).To(HavePrefix(`bug: expected dpkg list to contain at least one package from story 12345, but there were none`))
		})
	})
	When("story package is older than dpkg", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.2",
			}
			dpkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.3",
			}
		})
		It("is okay", func() {
			result, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})
	When("story package is equal to dpkg", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.2",
			}
			dpkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.2",
			}
		})
		It("is okay", func() {
			result, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})
	When("one story package is newer than dpkg", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.2",
				"gcc":        "6",
			}
			dpkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.1",
				"gcc":        "6",
			}
		})
		It("doesn't pass the check", func() {
			result, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ConsistOf("python-apt"))
		})
	})
	When("multiple packages are too old", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"python-apt":  "1.1.0~beta1ubuntu0.16.04.7",
				"python3-apt": "1.1.0~beta1ubuntu0.16.04.7",
			}
			dpkg = PackageVersions{
				"python-apt":  "1.0.0~beta1ubuntu0.16.04.7",
				"python3-apt": "1.0.0~beta1ubuntu0.16.04.7",
			}
		})
		It("doesn't pass the check", func() {
			result, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ConsistOf("python-apt", "python3-apt"))
		})
	})
	When("a package in a story is not parsable", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"python-apt": "@@@",
			}
			dpkg = PackageVersions{
				"python-apt": "1.9.0ubuntu1.2",
			}
		})
		It("returns an error", func() {
			_, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix(`parsing version from story for package "python-apt": `))
		})
	})
	When("a package in dpkg is not parsable", func() {
		BeforeEach(func() {
			storyPkg = PackageVersions{
				"python": "1.9.0ubuntu1.2",
			}
			dpkg = PackageVersions{
				"python": "@@@",
			}
		})
		It("returns an error", func() {
			_, err := FindPackagesRequiringUpdate(dpkg, story)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix(`parsing version from receipt for package "python": `))
		})
	})
})

var _ = Describe("ProcessSecurityStories", func() {
	var (
		dpkg    PackageVersions
		stories []*SecurityStory
	)
	BeforeEach(func() {
		b := strings.Builder{}
		b.WriteString("**USN**: https://usn.ubuntu.com/4233-1/\n")
		b.WriteString("**19.10 Packages:**\n")
		b.WriteString("python-apt 1.9.0ubuntu1.2\n")
		b.WriteString("python3-apt 1.9.0ubuntu1.2\n")
		b.WriteString("**18.04 Packages:**\n")
		b.WriteString("python-apt 1.1.0~beta1ubuntu0.16.04.7\n")
		b.WriteString("python3-apt 1.1.0~beta1ubuntu0.16.04.7\n")
		story1 := &SecurityStory{
			Title:       "**[Security Notice]** New USN affecting GPDB for Kubernetes: USN-4233-1: Python update",
			Description: b.String(),
		}

		b.Reset()
		b.WriteString("**USN**: https://usn.ubuntu.com/4234-1/\n")
		b.WriteString("**18.04 Packages:**\n")
		b.WriteString("dbus 1.10.6-1ubuntu3.5\n")
		story2 := &SecurityStory{
			Title:       "**[Security Notice]** New USN affecting GPDB for Kubernetes: USN-4234-1: DBUS update",
			Description: b.String(),
		}

		stories = []*SecurityStory{story1, story2}
	})
	var (
		updatesNeeded []string
		releaseNote   string
	)
	JustBeforeEach(func() {
		updatesNeeded, releaseNote = ProcessSecurityStories(stories, dpkg)
	})
	When("all package versions are ok", func() {
		BeforeEach(func() {
			dpkg = PackageVersions{
				"python-apt":  "1.1.0~beta1ubuntu0.16.04.7",
				"python3-apt": "1.1.0~beta1ubuntu0.16.04.7",
				"dbus":        "1.10.6-1ubuntu3.5",
			}
		})
		It("returns an empty list of packages needing updates", func() {
			Expect(updatesNeeded).To(BeEmpty())
		})
		It("produces release note", func() {
			Expect(releaseNote).To(HavePrefix("# Release Notes\nUpdated Ubuntu packages to address USNs:\n"))
			Expect(releaseNote).To(ContainSubstring("* [USN-4233-1](https://usn.ubuntu.com/4233-1/): Python update\n"))
			Expect(releaseNote).To(ContainSubstring("* [USN-4234-1](https://usn.ubuntu.com/4234-1/): DBUS update\n"))
		})
	})
	When("one story contains a package that is too old", func() {
		BeforeEach(func() {
			dpkg = PackageVersions{
				"python-apt":  "1.0.0~beta1ubuntu0.16.04.7",
				"python3-apt": "1.1.0~beta1ubuntu0.16.04.7",
				"dbus":        "1.10.6-1ubuntu3.5",
			}
		})
		It("returns the name of the bad package", func() {
			Expect(updatesNeeded).To(ConsistOf("python-apt"))
		})
		It("does not generate a release note for the outdated package", func() {
			Expect(releaseNote).To(BeEmpty())
		})
	})
	When("multiple stories contain packages that are too old", func() {
		BeforeEach(func() {
			dpkg = PackageVersions{
				"python-apt":  "1.0.0~beta1ubuntu0.16.04.7",
				"python3-apt": "1.1.0~beta1ubuntu0.16.04.7",
				"dbus":        "1.10.6-1ubuntu3.4",
			}
		})
		It("returns the names of all bad packages", func() {
			Expect(updatesNeeded).To(ConsistOf("python-apt", "dbus"))
		})
	})
})
