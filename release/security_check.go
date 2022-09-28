package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocarina/gocsv"
	version "github.com/knqyf263/go-deb-version"
)

var (
	storiesFn = flag.String("securityNotices", "", "CSV file with Security Notice stories exported from Tracker")
	receiptFn = flag.String("receipt", "greenplum-for-kubernetes-receipt.txt", "Text file with output of `dpkg -l` listing all of the packages in the release")
)

type SecurityStory struct {
	ID          string `csv:"Id"`
	Title       string `csv:"Title"`
	Description string `csv:"Description"`
	URL         string `csv:"URL"`

	usn struct {
		id       string
		link     string
		headline string
		packages PackageVersions
	}
}

func main() {
	flag.Parse()

	receiptFile, err := os.Open(*receiptFn)
	if err != nil {
		log.Fatal("open receipt: ", err)
	}
	receiptPkgs := ParsePackageReceipt(receiptFile)

	storiesFile, err := os.Open(*storiesFn)
	if err != nil {
		log.Fatal("open stories CSV: ", err)
	}

	var stories []*SecurityStory
	if err := gocsv.Unmarshal(storiesFile, &stories); err != nil {
		log.Fatal("unmarshal stories CSV: ", err)
	}

	pkgsRequiringUpdate, releaseNote := ProcessSecurityStories(stories, receiptPkgs)
	if len(pkgsRequiringUpdate) > 0 {
		log.Printf("packages that need to be updated: %v", pkgsRequiringUpdate)
	} else {
		fmt.Print(releaseNote)
	}
}

type PackageVersions map[string]string

func ParsePackageReceipt(receipt io.Reader) PackageVersions {
	packages := PackageVersions{}
	scanner := bufio.NewScanner(receipt)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		if len(words) < 4 || words[0] != "ii" {
			continue
		}
		pkg, ver := words[1], words[2]
		pkgParts := strings.Split(pkg, ":")
		packages[pkgParts[0]] = ver
	}
	return packages
}

var (
	usnRegexp     = regexp.MustCompile(`\b(?P<id>USN-\d+-\d+): (?P<headline>.*)`)
	usnLinkRegexp = regexp.MustCompile(`(?m)^\*\*USN\*\*: (?P<link>https://\S*)`)
)

func (ss *SecurityStory) ParseUSN() error {
	packages := PackageVersions{}
	scanner := bufio.NewScanner(strings.NewReader(ss.Description))
	ourPackages := false
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		switch {
		case line == "**18.04 Packages:**":
			ourPackages = true
		case strings.HasPrefix(line, "**"):
			ourPackages = false
		case ourPackages:
			fields := strings.Fields(line)
			packages[fields[0]] = fields[1]
		}
	}
	ss.usn.packages = packages

	if m := usnRegexp.FindStringSubmatch(ss.Title); len(m) == 3 {
		ss.usn.id = m[1]
		ss.usn.headline = m[2]
	} else {
		return fmt.Errorf("could not parse USN title and ID from story: %s", ss.ID)
	}

	if descMatch := usnLinkRegexp.FindStringSubmatch(ss.Description); len(descMatch) == 2 {
		ss.usn.link = descMatch[1]
	} else {
		return fmt.Errorf("could not parse USN link from story: %s", ss.ID)
	}

	return nil
}

func FindPackagesRequiringUpdate(dpkg PackageVersions, story SecurityStory) (result []string, err error) {
	storyPkgs := story.usn.packages
	matchedPkgs := 0
	for pkgName, storyVerStr := range storyPkgs {
		if sysVerStr, ok := dpkg[pkgName]; ok {
			matchedPkgs++
			var storyVersion, dpkgVersion version.Version
			if storyVersion, err = version.NewVersion(storyVerStr); err != nil {
				err = fmt.Errorf("parsing version from story for package %q: %w", pkgName, err)
				return
			}
			if dpkgVersion, err = version.NewVersion(sysVerStr); err != nil {
				err = fmt.Errorf("parsing version from receipt for package %q: %w", pkgName, err)
				return
			}
			if dpkgVersion.LessThan(storyVersion) {
				result = append(result, pkgName)
			}
		}
	}
	if matchedPkgs == 0 {
		err = fmt.Errorf("bug: expected dpkg list to contain at least one package from story %s, but there were none", story.ID)
	}
	return
}

func ProcessSecurityStories(stories []*SecurityStory, dpkg PackageVersions) (updatesNeeded []string, releaseNote string) {
	rn := strings.Builder{}
	rn.WriteString("# Release Notes\n")
	rn.WriteString("Updated Ubuntu packages to address USNs:\n")
	for _, story := range stories {
		if err := story.ParseUSN(); err != nil {
			log.Fatal(err)
		}
		pkgsRequiringUpdate, err := FindPackagesRequiringUpdate(dpkg, *story)
		if err != nil {
			log.Fatal(err)
		}
		updatesNeeded = append(updatesNeeded, pkgsRequiringUpdate...)
		rn.WriteString(GenerateReleaseNote(story))
	}
	if len(updatesNeeded) == 0 {
		releaseNote = rn.String()
	}
	return
}

func GenerateReleaseNote(story *SecurityStory) string {
	return fmt.Sprintf("* [%s](%s): %s\n", story.usn.id, story.usn.link, story.usn.headline)
}
