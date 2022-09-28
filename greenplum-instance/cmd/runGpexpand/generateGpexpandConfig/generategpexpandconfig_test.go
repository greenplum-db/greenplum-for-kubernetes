package gpexpandconfig

import (
	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/testing/matcher"
)

var _ = Describe("ExecPsqlQueryAndReturnInt", func() {
	var cmdFake *commandable.CommandFake
	BeforeEach(func() {
		cmdFake = commandable.NewFakeCommand()
	})
	When("query succeeds and result is an int", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc", "a fake query").
				ReturnsStatus(0).
				PrintsOutput("42\n")
		})
		It("returns the integer value", func() {
			resultInt, err := ExecPsqlQueryAndReturnInt(cmdFake.Command, "a fake query")
			Expect(err).NotTo(HaveOccurred())
			Expect(resultInt).To(Equal(42))
		})
	})
	When("there is an error running the query", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc", "a fake query").
				ReturnsStatus(1).
				PrintsError("psql failed")
		})
		It("returns error", func() {
			_, err := ExecPsqlQueryAndReturnInt(cmdFake.Command, "a fake query")
			Expect(err).To(MatchError("psql failed: exit status 1"))
		})
	})
	When("result is not an int", func() {
		BeforeEach(func() {
			cmdFake.FakeOutput("a string")
		})
		It("returns error", func() {
			_, err := ExecPsqlQueryAndReturnInt(cmdFake.Command, "a fake query")
			Expect(err).To(MatchError("strconv.Atoi: parsing \"a string\": invalid syntax"))
		})
	})
})

var _ = Describe("Run", func() {
	var (
		fs      vfs.Filesystem
		config  *GenerateGpexpandConfigParams
		cmdFake *commandable.CommandFake
	)
	BeforeEach(func() {
		fs = memfs.Create()
		cmdFake = commandable.NewFakeCommand()
		Expect(vfs.MkdirAll(fs, "/tmp", 0644)).To(Succeed())
		config = &GenerateGpexpandConfigParams{
			OldSegmentCount: 1,
			NewSegmentCount: 3,
			IsMirrored:      true,
			Fs:              fs,
			Command:         cmdFake.Command,
		}
		Expect(vfs.MkdirAll(fs, "/var/run/secrets/kubernetes.io/serviceaccount/", 0644)).To(Succeed())
		Expect(vfs.WriteFile(fs, "/var/run/secrets/kubernetes.io/serviceaccount/namespace", []byte("test-namespace"), 0777)).To(Succeed())
		cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc",
			"SELECT MAX(content) FROM gp_segment_configuration",
		).PrintsOutput("0\n")
		cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc",
			"SELECT MAX(dbid) FROM gp_segment_configuration",
		).PrintsOutput("4\n")
	})
	Describe("happy path", func() {
		When("mirrors=yes", func() {
			It("generates config successfully", func() {
				Expect(config.Run()).To(Succeed())
				Expect("/tmp/gpexpand_config").To(matcher.EqualInFilesystem(fs, `segment-a-1.agent.test-namespace.svc.cluster.local|segment-a-1|40000|/greenplum/data|5|1|p
segment-b-1.agent.test-namespace.svc.cluster.local|segment-b-1|50000|/greenplum/mirror/data|6|1|m
segment-a-2.agent.test-namespace.svc.cluster.local|segment-a-2|40000|/greenplum/data|7|2|p
segment-b-2.agent.test-namespace.svc.cluster.local|segment-b-2|50000|/greenplum/mirror/data|8|2|m
`))
			})
		})
		When("mirrors=no", func() {
			BeforeEach(func() {
				config.IsMirrored = false
			})
			It("generates config successfully", func() {
				Expect(config.Run()).To(Succeed())
				Expect("/tmp/gpexpand_config").To(matcher.EqualInFilesystem(fs, `segment-a-1.agent.test-namespace.svc.cluster.local|segment-a-1|40000|/greenplum/data|5|1|p
segment-a-2.agent.test-namespace.svc.cluster.local|segment-a-2|40000|/greenplum/data|6|2|p
`))
			})
		})
	})

	When("newSegmentCount <= 0", func() {
		BeforeEach(func() {
			config.NewSegmentCount = -1
		})
		It("returns error", func() {
			Expect(config.Run()).To(MatchError("new segment count cannot be <= 0"))
		})
	})

	When("getting max dbId fails", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc",
				"SELECT MAX(dbid) FROM gp_segment_configuration",
			).ReturnsStatus(1).PrintsError("custom get max dbId error")
		})
		It("returns error", func() {
			Expect(config.Run()).To(MatchError("custom get max dbId error: exit status 1"))
		})
	})

	When("getting max contentId fails", func() {
		BeforeEach(func() {
			cmdFake.ExpectCommand("/usr/local/greenplum-db/bin/psql", "-U", "gpadmin", "-tAc",
				"SELECT MAX(content) FROM gp_segment_configuration",
			).ReturnsStatus(1).PrintsError("custom get max contentId error")
		})
		It("returns error", func() {
			Expect(config.Run()).To(MatchError("custom get max contentId error: exit status 1"))
		})
	})

	When("getting namespace fails", func() {
		BeforeEach(func() {
			Expect(vfs.RemoveAll(fs, "/var/run/secrets/kubernetes.io/serviceaccount/")).To(Succeed())
		})
		It("returns error", func() {
			Expect(config.Run()).To(MatchError("open /var/run/secrets/kubernetes.io/serviceaccount/namespace: file does not exist"))
		})
	})

	Describe("GenerateConfig", func() {
		When("writing to /tmp/gpexpand_config fails", func() {
			BeforeEach(func() {
				Expect(vfs.RemoveAll(fs, "/tmp")).To(Succeed())
			})
			It("returns error", func() {
				Expect(config.Run()).To(MatchError("open /tmp/gpexpand_config: file does not exist"))
			})
		})

		When("newSegmentCount < OldSegmentCount", func() {
			BeforeEach(func() {
				config.OldSegmentCount = 4
			})
			It("returns error", func() {
				Expect(config.Run()).To(MatchError("newSegmentCount cannot be less than or equal to OldSegmentCount"))
			})
		})

	})
})
