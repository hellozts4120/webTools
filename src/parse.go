package main

import (
	"regexp"
	"strings"
)

// URL represents state for a parsed URL
// URL = scheme://opaque?query#fragment
// opaque = authority/path
// authority = userinfo@host:port
type URL struct {
	Input string

	Scheme   string
	Opaque   string
	Query    string
	Fragment string

	// Elements of Opaque
	Authority string
	Path      string

	// Elements of Authority
	UserInfo string
	Host     string
	Port     string
}

// RFC3986: https://www.ietf.org/rfc/rfc3986.txt
// URI scheme registry: http://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml
// TODO: Normalize method; See RFC3986 section 6.2.2 for normalization ref

func namedMatches(matches []string, r *regexp.Regexp) map[string]string {
	result := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if name == "" {
			continue
		}
		if i >= len(matches) {
			result[name] = ""
		} else {
			result[name] = matches[i]
		}
	}
	return result
}

func splitAuthorityFromPath(opaque string) (string, string) {
	r := regexp.MustCompile("((//)?(?P<authority>[^/]+))?(?P<path>/.*)?")
	matches := namedMatches(r.FindStringSubmatch(opaque), r)
	return matches["authority"], matches["path"]
}

func splitUserinfoHostPortFromAuthority(authority string) (string, string, string) {
	userinfo := ""
	if delimPos := strings.LastIndex(authority, "@"); delimPos != -1 {
		userinfo = authority[0:delimPos]
		authority = authority[delimPos+1:]
	}

	parts := []string{
		"(", "(\\[(?P<host6>[^\\]]+)\\])", "|", "(?P<host>[^:]+)", ")?", // host6 | host
		"(:(?P<port>[0-9]+))?",
	}

	r := regexp.MustCompile(strings.Join(parts, ""))
	matches := namedMatches(r.FindStringSubmatch(authority), r)
	if matches["host"] == "" {
		matches["host"] = matches["host6"]
	}

	return userinfo, matches["host"], matches["port"]
}

// Split splits an URL in to its main components (scheme, opaque, query, fragment)
func splitMainPart(url string) (string, string, string, string) {
	parts := []string{
		"^((?P<scheme>[^:\\.]+):)?", // scheme is required by RFC3986 (S3) but we are intentionally allowing it to be omitted for convenience
		"(?P<opaque>(//)?[^?#]+)",   // opaque-part
		"(\\?(?P<query>[^#]+))?",    // query
		"(#(?P<fragment>.*))?",      // fragment
	}

	r := regexp.MustCompile(strings.Join(parts, ""))
	matches := namedMatches(r.FindStringSubmatch(url), r)

	return matches["scheme"], matches["opaque"], matches["query"], matches["fragment"]
}

func setDefaultPort(port string, scheme string) (string) {
	if port != "" {
		return port
	}

	if scheme == "http" {
		return "80"
	}
	if scheme == "https" {
		return "443"
	}

	return port
}

// Parse splits an URL in to as many parts as it can
func Parse(url string) *URL {
	result := &URL{}
	result.Input = url
	result.Scheme, result.Opaque, result.Query, result.Fragment = splitMainPart(url)
	result.Authority, result.Path = splitAuthorityFromPath(result.Opaque)
	result.UserInfo, result.Host, result.Port = splitUserinfoHostPortFromAuthority(result.Authority)
	result.Port = setDefaultPort(result.Port, result.Scheme)
	return result
}
