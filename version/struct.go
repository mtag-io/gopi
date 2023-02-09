package version

// Version represents a single semantic version.
type Version struct {
	major, minor, patch uint64
	original            string
	meta                string
	pre                 string
}
