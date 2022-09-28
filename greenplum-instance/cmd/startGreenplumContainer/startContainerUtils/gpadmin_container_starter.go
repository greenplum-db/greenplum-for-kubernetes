package startContainerUtils

import (
	"fmt"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/fileutil"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/starter"
)

const bashrcPath = "/home/gpadmin/.bashrc"

type GpadminContainerStarter struct {
	*starter.App
}

func (s *GpadminContainerStarter) Run() error {
	for _, step := range []func() error{
		s.WriteContentsToBashrc,
		s.SetupSSHForGpadmin,
		s.CreateSymLink,
		s.CreatePsqlHistory,
		s.CreateMirrorDir,
	} {
		if err := step(); err != nil {
			return err
		}
	}

	return nil
}

func (s *GpadminContainerStarter) WriteContentsToBashrc() error {
	Log.Info("starting Greenplum Container")
	pxfFilename := "/etc/config/pxfServiceName"

	toInsert := "source /usr/local/greenplum-db/greenplum_path.sh\n"

	if _, err := s.Fs.Stat(pxfFilename); err == nil {
		b, err := vfs.ReadFile(s.Fs, pxfFilename)
		if err != nil {
			return fmt.Errorf("failed to read %s, was configMap mounted properly?", pxfFilename)
		}
		pxfServiceName := string(b)
		if pxfServiceName != "" {
			toInsert += "export PXF_HOST=" + pxfServiceName + "\n"
		}
	}

	fileWriter := fileutil.FileWriter{WritableFileSystem: s.Fs}
	return fileWriter.Insert(bashrcPath, toInsert)
}

func (s *GpadminContainerStarter) CreateSymLink() error {
	Log.Info("creating symlink for gpAdminLogs")
	if err := vfs.MkdirAll(s.Fs, "/greenplum/gpAdminLogs", 0755); err != nil {
		return err
	}
	return s.Fs.Symlink("/greenplum/gpAdminLogs", "/home/gpadmin/gpAdminLogs")
}

func (s *GpadminContainerStarter) CreateMirrorDir() error {
	Log.Info("creating mirror dir /greenplum/mirror")
	return vfs.MkdirAll(s.Fs, "/greenplum/mirror", 0755)
}

func (s *GpadminContainerStarter) CreatePsqlHistory() error {
	const filename = "/home/gpadmin/.psql_history"
	Log.Info("creating " + filename + " file")

	return vfs.WriteFile(s.Fs, filename, []byte(""), 0600)
}

func (s *GpadminContainerStarter) SetupSSHForGpadmin() error {
	Log.Info("setting up ssh for gpadmin")

	const LocalSSHDirPath = "/home/gpadmin/.ssh"
	err := vfs.MkdirAll(s.Fs, LocalSSHDirPath, 0711)
	if err != nil {
		return err
	}

	idrsaBytes, err := vfs.ReadFile(s.Fs, "/etc/ssh-key/id_rsa")
	if err != nil {
		return err
	}
	idrsaPubBytes, err := vfs.ReadFile(s.Fs, "/etc/ssh-key/id_rsa.pub")
	if err != nil {
		return err
	}

	for _, file := range []struct {
		filename string
		content  []byte
	}{
		{filename: LocalSSHDirPath + "/id_rsa", content: idrsaBytes},
		{filename: LocalSSHDirPath + "/id_rsa.pub", content: idrsaPubBytes},
		{filename: LocalSSHDirPath + "/authorized_keys", content: idrsaPubBytes},
		{filename: LocalSSHDirPath + "/known_hosts", content: []byte{}},
		{filename: LocalSSHDirPath + "/config", content: []byte("Host *\n    ConnectionAttempts 5")},
	} {
		if err := vfs.WriteFile(s.Fs, file.filename, file.content, 0600); err != nil {
			return err
		}
	}

	return nil
}
