package net

import "fmt"

func GenerateHostList(segmentCount int, useMirrors, useStandby bool, dnsSuffix string) []string {
	hostList := generateHostListWithDNSSuffix(segmentCount, useMirrors, useStandby, "")
	if dnsSuffix != "" {
		hostList = append(hostList, generateHostListWithDNSSuffix(segmentCount, useMirrors, useStandby, dnsSuffix)...)
	}
	return hostList
}

func generateHostListWithDNSSuffix(segmentCount int, useMirrors, useStandby bool, dnsSuffix string) []string {
	var hostnames []string
	for i := 0; i < segmentCount; i++ {
		hostnames = append(hostnames, fmt.Sprintf("segment-a-%d%s", i, dnsSuffix))
		if useMirrors {
			hostnames = append(hostnames, fmt.Sprintf("segment-b-%d%s", i, dnsSuffix))
		}
	}
	hostnames = append(hostnames, fmt.Sprintf("master-0%s", dnsSuffix))
	if useStandby {
		hostnames = append(hostnames, fmt.Sprintf("master-1%s", dnsSuffix))
	}
	return hostnames
}
