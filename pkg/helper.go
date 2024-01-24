package pkg

import (
	"regexp"
	"strings"
)

type SortableVersions []string

func (a SortableVersions) Len() int      { return len(a) }
func (a SortableVersions) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortableVersions) Less(i, j int) bool {
	vi, vj := strings.TrimLeft(a[i], "v"), strings.TrimLeft(a[j], "v")
	major := regexp.MustCompile("^[0-9]+")
	viMajor, vjMajor := major.FindString(vi), major.FindString(vj)
	viRemaining, vjRemaining := strings.TrimLeft(vi, viMajor), strings.TrimLeft(vj, vjMajor)
	switch {
	case len(viRemaining) == 0 && len(vjRemaining) == 0:
		return viMajor < vjMajor
	case len(viRemaining) == 0 && len(vjRemaining) != 0:
		// stable version is greater than unstable version
		return false
	case len(viRemaining) != 0 && len(vjRemaining) == 0:
		// stable version is greater than unstable version
		return true
	}
	// neither are stable versions
	if viMajor != vjMajor {
		return viMajor < vjMajor
	}
	// assuming at most we have one alpha or one beta version, so if vi contains "alpha", it's the lesser one.
	return strings.Contains(viRemaining, "alpha")
}
