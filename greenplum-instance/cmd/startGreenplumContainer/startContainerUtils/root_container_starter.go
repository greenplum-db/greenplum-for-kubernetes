package startContainerUtils

import (
	"bufio"
	"strings"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/ubuntuUtils"
	"github.com/pkg/errors"
)

const HostKeyDir = "/greenplum/hostKeyDir"

type RootContainerStarter struct {
	*starter.App
	Ubuntu ubuntuUtils.UbuntuInterface
}

func (s *RootContainerStarter) Run() error {
	for _, step := range []func() error{
		s.CreateGpdbCgroup,
		s.ChownGreenplumDir,
		s.SetupSSHHostKeys,
		s.AddSubDomain,
	} {
		if err := step(); err != nil {
			return err
		}
	}

	return nil
}

func (s *RootContainerStarter) ChownGreenplumDir() error {
	Log.Info("changing ownership of /greenplum to gpadmin")
	err := s.Ubuntu.ChangeDirectoryOwner("/greenplum", "gpadmin")
	return errors.Wrap(err, "changing ownership of /greenplum dir to gpadmin failed")
}

func (s *RootContainerStarter) SetupSSHHostKeys() error {
	const sshHostRSAKeyPath = HostKeyDir + "/ssh_host_rsa_key"

	_, err := s.Fs.Lstat(sshHostRSAKeyPath)
	if err != nil {
		vfs.MkdirAll(s.Fs, HostKeyDir, 0755)
		cmd := s.Command("/usr/bin/ssh-keygen", "-t", "rsa", "-f", sshHostRSAKeyPath, "-N", "")
		cmd.Stderr = s.StderrBuffer
		err = cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "failed to generate SSH host key at %v ", sshHostRSAKeyPath)
		}
	}
	err = fileutil.CopyFile(s.Fs, sshHostRSAKeyPath, "/etc/ssh/ssh_host_rsa_key")
	return errors.Wrapf(err, "failed to copy SSH host key to /etc/ssh/ssh_host_rsa_key")
}

func (s *RootContainerStarter) CreateGpdbCgroup() error {

	Log.Info("Creating cgroup dirs")
	const selfCgroupFn = "/proc/self/cgroup"
	gpdbSubSystemNames := map[string]bool{
		"cpu":     true,
		"cpuacct": true,
		"cpuset":  true,
		"memory":  true,
	}

	var err error
	procMountMap, err := s.ParseProcCgroupMounts(gpdbSubSystemNames)
	if err != nil {
		return err
	}
	selfCgroupFile, err := vfs.Open(s.Fs, selfCgroupFn)
	if err != nil {
		return err
	}
	defer selfCgroupFile.Close()

	scanner := bufio.NewScanner(selfCgroupFile)
	for scanner.Scan() {
		lineContents := strings.Split(scanner.Text(), ":")
		subSystemNames := strings.Split(lineContents[1], ",")
		for _, subSystemName := range subSystemNames {
			if subsysMount, ok := procMountMap[subSystemName]; ok {
				dirPath := subsysMount + lineContents[2] + "/gpdb"
				cmd := s.Command("install", "-d", "-o", "gpadmin", "-g", "gpadmin", dirPath)
				cmd.Stdout = s.StdoutBuffer
				cmd.Stderr = s.StderrBuffer
				if err := cmd.Run(); err != nil {
					return errors.Wrapf(err, "failed to create cgroup dir %v with gpadmin as owner", dirPath)
				}
				// chown all the kernel-created files within the cgroup dir
				cmd = s.Command("chown", "-R", "gpadmin:gpadmin", dirPath)
				cmd.Stdout = s.StdoutBuffer
				cmd.Stderr = s.StderrBuffer
				if err := cmd.Run(); err != nil {
					return errors.Wrapf(err, "failed to change ownership of %v to gpadmin", dirPath)
				}
			}
		}
	}

	return nil
}

func (s *RootContainerStarter) ParseProcCgroupMounts(gpdbSubSystemNames map[string]bool) (map[string]string, error) {
	const procMountsFn = "/proc/mounts"
	procMountsFile, err := vfs.Open(s.Fs, procMountsFn)
	if err != nil {
		return nil, err
	}
	defer procMountsFile.Close()
	procMountScanner := bufio.NewScanner(procMountsFile)
	procMountMap := map[string]string{}
	for procMountScanner.Scan() {
		procMount := strings.Split(procMountScanner.Text(), " ")
		mountType := procMount[2]
		if mountType != "cgroup" {
			continue
		}
		mountOptions := strings.Split(procMount[3], ",")
		for _, mountOption := range mountOptions {
			if gpdbSubSystemNames[mountOption] {
				procMountMap[mountOption] = procMount[1]
			}
		}
	}
	return procMountMap, nil
}

func (s *RootContainerStarter) AddSubDomain() error {
	const resolvConfFilePath = "/etc/resolv.conf"
	fileContents, err := vfs.ReadFile(s.Fs, resolvConfFilePath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(fileContents), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "search") {
			searchDomains := strings.Split(line, " ")[1:]
			searchDomains = addAgentDomain(searchDomains)
			lines[i] = "search " + strings.Join(searchDomains, " ")
		}
	}

	output := strings.Join(lines, "\n")
	return vfs.WriteFile(s.Fs, resolvConfFilePath, []byte(output), 0644)
}

func addAgentDomain(domains []string) []string {
	namespaceDomain := getNamespaceDomain(domains)
	agentDomain := "agent." + namespaceDomain
	newDomains := make([]string, 1, len(domains)+1)
	newDomains[0] = agentDomain
	for _, domain := range domains {
		if domain == agentDomain {
			continue
		}
		newDomains = append(newDomains, domain)
	}
	return newDomains
}

func getNamespaceDomain(domains []string) string {
	const svcDomain = ".svc.cluster.local"
	for _, domain := range domains {
		namespace := strings.TrimSuffix(domain, svcDomain)
		if namespace != domain { // it had the suffix
			if !strings.Contains(namespace, ".") {
				return domain
			}

		}
	}
	return ""
}
