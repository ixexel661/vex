package xml

import "strings"

const whitespace = " \t\r\n"

// splitTag cuts a start tag like <a x="1" y = '2'/> into its name, its
// attributes like x="1" and its end, ">" or "/>".
func splitTag(tag string) (name string, attrs []string, end string) {
	end = ">"
	if strings.HasSuffix(tag, "/>") {
		end = "/>"
	}
	rest := strings.TrimSuffix(strings.TrimPrefix(tag, "<"), end)
	name, rest = cutName(rest)
	for {
		rest = strings.TrimLeft(rest, whitespace)
		if rest == "" {
			return name, attrs, end
		}
		var attr string
		attr, rest = cutAttribute(rest)
		attrs = append(attrs, attr)
	}
}

func cutName(s string) (name, rest string) {
	i := strings.IndexAny(s, whitespace)
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i:]
}

// cutAttribute cuts the attribute at the start of s and drops the spaces
// around its "=". The value keeps its quotes, so it may contain spaces and ">".
func cutAttribute(s string) (attr, rest string) {
	name, value, ok := strings.Cut(s, "=")
	value = strings.TrimLeft(value, whitespace)
	if !ok || value == "" {
		return s, ""
	}
	end := strings.IndexByte(value[1:], value[0]) + 2
	if end < 2 {
		return s, ""
	}
	return strings.TrimRight(name, whitespace) + "=" + value[:end], value[end:]
}

// selfClosing turns the start tag of an empty element into a self-closing
// tag, e.g. <a x="1"> into <a x="1"/>.
func selfClosing(tag string) string {
	if strings.HasSuffix(tag, "/>") {
		return tag
	}
	return strings.TrimRight(strings.TrimSuffix(tag, ">"), whitespace) + "/>"
}
