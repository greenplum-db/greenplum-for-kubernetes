package testing

import (
	"encoding/json"
	"io"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
)

func DecodeLogs(buf io.Reader) (logs []map[string]interface{}, _ error) {
	d := json.NewDecoder(buf)
	for {
		entry := map[string]interface{}{}
		err := d.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}
	return logs, nil
}

func ContainLogEntry(keys gstruct.Keys) types.GomegaMatcher {
	return gomega.ContainElement(gstruct.MatchKeys(gstruct.IgnoreExtras, keys))
}
