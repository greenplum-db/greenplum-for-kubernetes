local specialChars = "\\.+*?()|[]{}^$";

// escape_regex() escapes regular expression meta characters. The resulting regex will match the raw input string.
function(raw)
    std.join(
        "",
        std.flatMap(
            function(x)
                if std.member(std.stringChars(specialChars), x)
                then ['\\', x]
                else [x],
            std.stringChars(raw)))
