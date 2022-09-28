
local tests=[
    import "concourse_tests.jsonnet",
    import "escape_regex_tests.jsonnet",
];

local failing_tests = [ t for t in std.flattenArrays(tests) if !t.result ];

if std.length(failing_tests) == 0 then
    "PASS"
else
{
    failures: failing_tests
}
