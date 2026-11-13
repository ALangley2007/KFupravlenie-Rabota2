package cargo

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	versionComponents = 3
	orderBefore       = -1
	orderEqual        = 0
	orderAfter        = 1
)

var (
	numericIdentifier = regexp.MustCompile(`^[0-9]+$`)
	versionIdentifier = regexp.MustCompile(`^[0-9A-Za-z-]+$`)
)

type version struct {
	core       [versionComponents]string
	prerelease []string
}

func parseVersion(value string) (version, error) {
	core, build, hasBuild := strings.Cut(value, "+")
	if hasBuild && !validIdentifiers(build, false) {
		return version{}, fmt.Errorf("некорректная SemVer-версия %q", value)
	}
	core, prerelease, hasPrerelease := strings.Cut(core, "-")
	parts := strings.Split(core, ".")
	if len(parts) != versionComponents || (hasPrerelease && !validIdentifiers(prerelease, true)) {
		return version{}, fmt.Errorf("некорректная SemVer-версия %q", value)
	}
	var parsed version
	for i, part := range parts {
		if !validNumber(part) {
			return version{}, fmt.Errorf("некорректная SemVer-версия %q", value)
		}
		parsed.core[i] = part
	}
	if hasPrerelease {
		parsed.prerelease = strings.Split(prerelease, ".")
	}
	return parsed, nil
}

func validNumber(value string) bool {
	return numericIdentifier.MatchString(value) && (value == "0" || !strings.HasPrefix(value, "0"))
}

func validIdentifiers(value string, prerelease bool) bool {
	for _, identifier := range strings.Split(value, ".") {
		if !versionIdentifier.MatchString(identifier) {
			return false
		}
		if prerelease && numericIdentifier.MatchString(identifier) && !validNumber(identifier) {
			return false
		}
	}
	return true
}

func preferVersion(candidate, current version) bool {
	if (candidate.prerelease == nil) != (current.prerelease == nil) {
		return candidate.prerelease == nil
	}
	for i, part := range candidate.core {
		if order := compareNumber(part, current.core[i]); order != orderEqual {
			return order == orderAfter
		}
	}
	return comparePrerelease(candidate.prerelease, current.prerelease) == orderAfter
}

func compareNumber(left, right string) int {
	if len(left) < len(right) {
		return orderBefore
	}
	if len(left) > len(right) {
		return orderAfter
	}
	return strings.Compare(left, right)
}

func comparePrerelease(left, right []string) int {
	for i, identifier := range left {
		if i >= len(right) {
			return orderAfter
		}
		if order := compareIdentifier(identifier, right[i]); order != orderEqual {
			return order
		}
	}
	if len(left) < len(right) {
		return orderBefore
	}
	return orderEqual
}

func compareIdentifier(left, right string) int {
	leftNumber, rightNumber := numericIdentifier.MatchString(left), numericIdentifier.MatchString(right)
	if leftNumber && rightNumber {
		return compareNumber(left, right)
	}
	if leftNumber != rightNumber {
		if leftNumber {
			return orderBefore
		}
		return orderAfter
	}
	return strings.Compare(left, right)
}
