package update

import (
	"regexp"
	"strings"
)

// Accept full SemVer only. Numeric identifiers are compared as strings so a
// long (but valid) release identifier cannot overflow a machine integer.
var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

type version struct {
	core []string
	pre  []string
}

func parseVersion(s string) (version, bool) {
	if len(s) > 128 {
		return version{}, false
	}
	parts := versionPattern.FindStringSubmatch(s)
	if parts == nil {
		return version{}, false
	}
	v := version{core: parts[1:4]}
	if parts[4] == "" {
		return v, true
	}
	v.pre = strings.Split(parts[4], ".")
	for _, id := range v.pre {
		if len(id) > 1 && id[0] == '0' && numeric(id) {
			return version{}, false
		}
	}
	return v, true
}

func numeric(id string) bool { return strings.Trim(id, "0123456789") == "" }

func compareNumber(a, b string) int {
	if len(a) != len(b) {
		return len(a) - len(b)
	}
	return strings.Compare(a, b)
}

func compareID(a, b string) int {
	switch {
	case numeric(a) && numeric(b):
		return compareNumber(a, b)
	case numeric(a):
		return -1
	case numeric(b):
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func (v version) compare(other version) int {
	for i, part := range v.core {
		if order := compareNumber(part, other.core[i]); order != 0 {
			return order
		}
	}
	if len(v.pre) == 0 || len(other.pre) == 0 {
		return len(other.pre) - len(v.pre)
	}
	for i := 0; i < min(len(v.pre), len(other.pre)); i++ {
		if order := compareID(v.pre[i], other.pre[i]); order != 0 {
			return order
		}
	}
	return len(v.pre) - len(other.pre)
}
