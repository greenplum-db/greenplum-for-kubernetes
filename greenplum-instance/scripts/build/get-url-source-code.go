package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// INVALIDURL if URL is not found
const INVALIDURL = "NO VALID URL"

// FILESUFFIXES list of file suffixes to check in the ftp server
var FILESUFFIXES = []string{"_amd64.tar.gz", ".tar.gz", ".tar.xz", ".debian.tar.xz"}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "source.csv out.csv")
		os.Exit(1)
	}

	// Read from the file
	// format: binaryName,version
	readFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("error opening 'a'\n")
	}
	defer func() {
		readFileErr := readFile.Close()
		if readFileErr != nil {
			log.Fatal(readFileErr)
		}
	}()

	// Write to the file
	// format: cleanupbinaryName,cleanedupversion,url
	writeFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatalf("error writing 'a'\n")
	}
	defer func() {
		writeFileErr := writeFile.Close()
		if writeFileErr != nil {
			log.Fatal(writeFileErr)
		}
	}()

	scanner := bufio.NewScanner(readFile)
	writer := bufio.NewWriter(writeFile)

	for scanner.Scan() {
		binaryText := scanner.Text()
		splitString := strings.Split(binaryText, ",")
		binaryName := cleanupBinaryName(splitString[0])
		version := cleanupVersion(splitString[1])
		category := getCategory(binaryName)
		url := findValidURL(fmt.Sprintf("http://archive.ubuntu.com/ubuntu/pool/main/%s/%s/%s_%s", category, binaryName, binaryName, version))
		if url == INVALIDURL {
			url = getOtherWorkingURL(binaryName)
		}
		outputLine := fmt.Sprintf("%s,%s,%s", binaryName, version, url)
		fmt.Println(outputLine)
		fmt.Fprintln(writer, outputLine)
		flushErr := writer.Flush()
		if flushErr != nil {
			fmt.Println("Flush error")
			os.Exit(1)
		}
	}
	flushErr := writer.Flush()
	if flushErr != nil {
		fmt.Println("Flush error")
		os.Exit(1)
	}
}

func cleanupVersion(version string) string {
	s := strings.Split(version, ":")
	if len(s) > 1 {
		version = s[1]
	}
	return strings.TrimRight(version, ".")
}

func cleanupBinaryName(binaryName string) string {
	s := strings.Split(binaryName, ":")
	if len(s) > 1 {
		return s[0]
	}
	return binaryName
}

func findValidURL(prefixURL string) string {
	for _, suffix := range FILESUFFIXES {
		if isURLValid(fmt.Sprintf("%s%s", prefixURL, suffix)) {
			return fmt.Sprintf("%s%s", prefixURL, suffix)
		}
	}
	return INVALIDURL
}

func isURLValid(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func getCategory(binaryName string) string {
	if strings.HasPrefix(binaryName, "lib") {
		return string(binaryName[0:4])
	}
	return string(binaryName[0])
}

func getOtherWorkingURL(binaryName string) string {
	ubuntuPackageURL := fmt.Sprintf("https://packages.ubuntu.com/xenial/amd64/%s/download", binaryName)
	resp, _ := http.Get(ubuntuPackageURL)
	if resp.StatusCode == 200 {
		htmlBytes, _ := ioutil.ReadAll(resp.Body)
		html := string(htmlBytes)
		r := regexp.MustCompile(`You can download the requested file from the <tt>(?P<dir>.*)</tt>`)
		match := r.FindStringSubmatch(html)
		return fmt.Sprintf("http://archive.ubuntu.com/ubuntu/%s", match[1])
	}
	return INVALIDURL
}
