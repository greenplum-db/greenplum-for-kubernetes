package cluster_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-instance/cmd/startGreenplumContainer/startContainerUtils/cluster"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

var _ = Describe("greenplum_command", func() {
	var (
		subject *cluster.GreenplumCommand
	)

	When("NewGreenplumCommand is called", func() {
		It("creates a GreenplumCommand", func() {
			subject = cluster.NewGreenplumCommand(commandable.NewFakeCommand().Command)
			Expect(subject).NotTo(BeNil())
		})
	})

	When("GreenplumCommand.Command() is called", func() {
		It("sets the proper environment variables", func() {
			subject = cluster.NewGreenplumCommand(commandable.NewFakeCommand().Command)
			cmd := subject.Command("foo", "bar")
			Expect(cmd.Env).To(ContainElement("HOME=/home/gpadmin"))
			Expect(cmd.Env).To(ContainElement("USER=gpadmin"))
			Expect(cmd.Env).To(ContainElement("LOGNAME=gpadmin"))
			Expect(cmd.Env).To(ContainElement("GPHOME=/usr/local/greenplum-db"))
			Expect(cmd.Env).To(ContainElement("PATH=/usr/local/greenplum-db/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"))
			Expect(cmd.Env).To(ContainElement("LD_LIBRARY_PATH=/usr/local/greenplum-db/lib:/usr/local/greenplum-db/ext/python/lib"))
			Expect(cmd.Env).To(ContainElement("MASTER_DATA_DIRECTORY=/greenplum/data-1"))
			Expect(cmd.Env).To(ContainElement("PYTHONPATH=/usr/local/greenplum-db/lib/python"))
		})
	})
})
