package instanceconfig

import (
	"fmt"
	"os"
	"strconv"

	"github.com/blang/vfs"
)

const ConfigMapPathPrefix = "/etc/config/"
const PodInfoPathPrefix = "/etc/podinfo/"

type ConfigValues struct {
	Namespace            string
	GreenplumClusterName string
	SegmentCount         int
	Mirrors              bool
	Standby              bool
	PXFServiceName       string
}

type Reader interface {
	GetNamespace() (string, error)
	GetGreenplumClusterName() (string, error)
	GetSegmentCount() (int, error)
	GetMirrors() (bool, error)
	GetStandby() (bool, error)
	GetPXFServiceName() (string, error)
	GetConfigValues() (ConfigValues, error)
}

type fsReader struct {
	fs vfs.Filesystem
}

var _ Reader = &fsReader{}

func NewReader(fs vfs.Filesystem) Reader {
	return &fsReader{fs: fs}
}

func (cr *fsReader) GetNamespace() (string, error) {
	return cr.readString(PodInfoPathPrefix, "namespace")
}

func (cr *fsReader) GetGreenplumClusterName() (string, error) {
	return cr.readString(PodInfoPathPrefix, "greenplumClusterName")
}

func (cr *fsReader) GetSegmentCount() (int, error) {
	return cr.readInt(ConfigMapPathPrefix, "segmentCount")
}

func (cr *fsReader) GetMirrors() (bool, error) {
	return cr.readBool(ConfigMapPathPrefix, "mirrors")
}

func (cr *fsReader) GetStandby() (bool, error) {
	return cr.readBool(ConfigMapPathPrefix, "standby")
}

func (cr *fsReader) GetPXFServiceName() (string, error) {
	return cr.readOptionalString(ConfigMapPathPrefix, "pxfServiceName")
}

func (cr *fsReader) GetConfigValues() (ConfigValues, error) {
	configValues := ConfigValues{}
	var err error

	configValues.Namespace, err = cr.GetNamespace()
	if err != nil {
		return ConfigValues{}, err
	}

	configValues.GreenplumClusterName, err = cr.GetGreenplumClusterName()
	if err != nil {
		return ConfigValues{}, err
	}

	configValues.SegmentCount, err = cr.GetSegmentCount()
	if err != nil {
		return ConfigValues{}, err
	}

	configValues.Mirrors, err = cr.GetMirrors()
	if err != nil {
		return ConfigValues{}, err
	}

	configValues.Standby, err = cr.GetStandby()
	if err != nil {
		return ConfigValues{}, err
	}

	configValues.PXFServiceName, err = cr.GetPXFServiceName()
	if err != nil {
		return ConfigValues{}, err
	}

	return configValues, nil

}

func (cr *fsReader) readInt(pathPrefix, configKey string) (value int, err error) {
	valueBytes, err := vfs.ReadFile(cr.fs, pathPrefix+configKey)
	if err != nil {
		return 0, err
	}
	value, err = strconv.Atoi(string(valueBytes))
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		err := fmt.Errorf("%s must be > 0", configKey)
		return 0, err
	}
	return value, nil
}

func (cr *fsReader) readBool(pathPrefix, configKey string) (bool, error) {
	configFilename := pathPrefix + configKey
	valueBytes, err := vfs.ReadFile(cr.fs, configFilename)
	if err != nil {
		return false, fmt.Errorf("error reading %s: %s", configKey, err.Error())
	}

	value, err := strconv.ParseBool(string(valueBytes))
	if err != nil {
		return false, fmt.Errorf("error parsing %s, must be a boolean, got: %s", configKey, string(valueBytes))
	}
	return value, nil
}

func (cr *fsReader) readString(pathPrefix, configKey string) (string, error) {
	return cr.readStringHelper(pathPrefix, configKey, true)
}

func (cr *fsReader) readOptionalString(pathPrefix, configKey string) (string, error) {
	return cr.readStringHelper(pathPrefix, configKey, false)
}

func (cr *fsReader) readStringHelper(pathPrefix, configKey string, required bool) (string, error) {
	configFilename := pathPrefix + configKey
	b, err := vfs.ReadFile(cr.fs, configFilename)
	if err != nil {
		if !required && os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}
