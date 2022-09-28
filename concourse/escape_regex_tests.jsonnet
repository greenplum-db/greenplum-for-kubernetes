local escape_regex = import "escape_regex.libsonnet";

[
    {
        description: "Leaves simple strings alone",
        expected: "Hello, world!",
        actual: escape_regex("Hello, world!"),
        result: self.expected == self.actual,
    },
    {
        description: "A version string",
        expected: "2\\.0\\.0-alpha\\.1",
        actual: escape_regex("2.0.0-alpha.1"),
        result: self.expected == self.actual,
    },
    {
        description: "All the regex meta characters",
        expected: "\\\\" + "\\." + "\\+" + "\\*" + "\\?" + "\\(" + "\\)" + "\\|" + "\\[" + "\\]" + "\\{" + "\\}" + "\\^" + "\\$",
        actual: escape_regex("\\.+*?()|[]{}^$"),  // Copied from Golang's regexp.QuoteMeta() https://golang.org/src/regexp/regexp.go
        result: self.expected == self.actual,
    },
]
