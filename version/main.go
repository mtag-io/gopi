package version

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// ClassRegex The compiled Class of the regex created at init() is cached here so it
// only needs to be created once.
var ClassRegex *regexp.Regexp

var (
	// ErrInvalidSemVer is returned a Class is found to be invalid when
	// being parsed.
	ErrInvalidSemVer = errors.New("invalid Semantic Class")

	// ErrEmptyString is returned when an empty string is passed in for parsing.
	ErrEmptyString = errors.New("class string empty")

	// ErrInvalidCharacters is returned when invalid characters are found as
	// part of a Class
	ErrInvalidCharacters = errors.New("invalid characters in Class")

	// ErrSegmentStartsZero is returned when a Class segment starts with 0.
	// This is invalid in SemVer.
	ErrSegmentStartsZero = errors.New("class segment starts with 0")

	// ErrInvalidMetadata is returned when the metadata is an invalid format
	ErrInvalidMetadata = errors.New("invalid Metadata string")

	// ErrInvalidPrerelease is returned when the pre-release is an invalid format
	ErrInvalidPrerelease = errors.New("invalid Prerelease string")
)

// semVerRegex is the regular expression used to parse a semantic Class.
const semVerRegex string = `v?([0-9]+)(\.[0-9]+)?(\.[0-9]+)?` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

// Class represents a single semantic Class.
type Class struct {
	major, minor, patch uint64
	pre                 string
	metadata            string
	original            string
}

func init() {
	ClassRegex = regexp.MustCompile("^" + semVerRegex + "$")
}

const (
	num     string = "0123456789"
	allowed        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-" + num
)

// StrictNew parses a given Class and returns an instance of Class or
// an error if unable to parse the Class. Only parses valid semantic Class.
// Performs checking that can find errors within the Class.
// If you want to coerce a Class such as 1 or 1.2 and parse it as the 1.x
// releases of semver did, use the New() function.
func StrictNew(v string) (*Class, error) {
	// Parsing here does not use RegEx in order to increase performance and reduce
	// allocations.

	if len(v) == 0 {
		return nil, ErrEmptyString
	}

	// Split the parts into [0]major, [1]minor, and [2]patch,prerelease,build
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil, ErrInvalidSemVer
	}

	sv := &Class{
		original: v,
	}

	// check for prerelease or build metadata
	var extra []string
	if strings.ContainsAny(parts[2], "-+") {
		// Start with the build metadata first as it needs to be on the right
		extra = strings.SplitN(parts[2], "+", 2)
		if len(extra) > 1 {
			// build metadata found
			sv.metadata = extra[1]
			parts[2] = extra[0]
		}

		extra = strings.SplitN(parts[2], "-", 2)
		if len(extra) > 1 {
			// prerelease found
			sv.pre = extra[1]
			parts[2] = extra[0]
		}
	}

	// Validate the number segments are valid. This includes only having positive
	// numbers and no leading 0's.
	for _, p := range parts {
		if !containsOnly(p, num) {
			return nil, ErrInvalidCharacters
		}

		if len(p) > 1 && p[0] == '0' {
			return nil, ErrSegmentStartsZero
		}
	}

	// Extract the major, minor, and patch elements onto the returned Class
	var err error
	sv.major, err = strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}

	sv.minor, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	sv.patch, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, err
	}

	// No prerelease or build metadata found so returning now as a fastpath.
	if sv.pre == "" && sv.metadata == "" {
		return sv, nil
	}

	if sv.pre != "" {
		if err = validatePrerelease(sv.pre); err != nil {
			return nil, err
		}
	}

	if sv.metadata != "" {
		if err = validateMetadata(sv.metadata); err != nil {
			return nil, err
		}
	}

	return sv, nil
}

// New parses a given Class and returns an instance of Class or
// an error if unable to parse the Class. If the Class is SemVer-ish it
// attempts to convert it to SemVer. If you want  to validate it was a strict
// semantic Class at parse time see StrictNew().
func New(v string) (*Class, error) {
	m := ClassRegex.FindStringSubmatch(v)
	if m == nil {
		return nil, ErrInvalidSemVer
	}

	sv := &Class{
		metadata: m[8],
		pre:      m[5],
		original: v,
	}

	var err error
	sv.major, err = strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing Class segment: %s", err)
	}

	if m[2] != "" {
		sv.minor, err = strconv.ParseUint(strings.TrimPrefix(m[2], "."), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing Class segment: %s", err)
		}
	} else {
		sv.minor = 0
	}

	if m[3] != "" {
		sv.patch, err = strconv.ParseUint(strings.TrimPrefix(m[3], "."), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing Class segment: %s", err)
		}
	} else {
		sv.patch = 0
	}

	// Perform some basic due diligence on the extra parts to ensure they are
	// valid.

	if sv.pre != "" {
		if err = validatePrerelease(sv.pre); err != nil {
			return nil, err
		}
	}

	if sv.metadata != "" {
		if err = validateMetadata(sv.metadata); err != nil {
			return nil, err
		}
	}

	return sv, nil
}

// NewByArgs creates a new instance of Class with each of the parts passed in as
// arguments instead of parsing a Class string.
func NewByArgs(major, minor, patch uint64, pre, metadata string) *Class {
	v := Class{
		major:    major,
		minor:    minor,
		patch:    patch,
		pre:      pre,
		metadata: metadata,
		original: "",
	}

	v.original = v.String()

	return &v
}

// MustParse parses a given Class and panics on error.
func MustParse(v string) *Class {
	sv, err := New(v)
	if err != nil {
		panic(err)
	}
	return sv
}

// String converts a Class object to a string.
// Note, if the original Class contained a leading v this Class will not.
// See the Original() method to retrieve the original value. Semantic Class
// don't contain a leading v per the spec. Instead it's optional on
// implementation.
func (v *Class) String() string {
	var buf bytes.Buffer

	_, err := fmt.Fprintf(&buf, "%d.%d.%d", v.major, v.minor, v.patch)
	if err != nil {
		return ""
	}
	if v.pre != "" {
		_, err = fmt.Fprintf(&buf, "-%s", v.pre)
		if err != nil {
			log.Fatalln("Unable to parse semver")
		}
	}
	if v.metadata != "" {
		_, err = fmt.Fprintf(&buf, "+%s", v.metadata)
		if err != nil {
			log.Fatalln("Unable to parse semver")
		}
	}

	return buf.String()
}

// Original returns the original value passed in to be parsed.
func (v *Class) Original() string {
	return v.original
}

// Major returns the major Class.
func (v *Class) Major() uint64 {
	return v.major
}

// Minor returns the minor Class.
func (v *Class) Minor() uint64 {
	return v.minor
}

// Patch returns the patch Class.
func (v *Class) Patch() uint64 {
	return v.patch
}

// Prerelease returns the prerelease Class.
func (v *Class) Prerelease() string {
	return v.pre
}

// Metadata returns the metadata on the Class.
func (v *Class) Metadata() string {
	return v.metadata
}

// originalVPrefix returns the original 'v' prefix if any.
func (v *Class) originalVPrefix() string {
	// Note, only lowercase v is supported as a prefix by the parser.
	if v.original != "" && v.original[:1] == "v" {
		return v.original[:1]
	}
	return ""
}

// IncPatch produces the next patch Class.
// If the current Class does not have prerelease/metadata information,
// it unsets metadata and prerelease values, increments patch number.
// If the current Class has any of prerelease or metadata information,
// it unsets both values and keeps current patch value
func (v *Class) IncPatch() Class {
	vNext := v
	// according to http://semver.org/#spec-item-9
	// Pre-release Class have a lower precedence than the associated normal Class.
	// according to http://semver.org/#spec-item-10
	// Build metadata SHOULD be ignored when determining Class precedence.
	if v.pre != "" {
		vNext.metadata = ""
		vNext.pre = ""
	} else {
		vNext.metadata = ""
		vNext.pre = ""
		vNext.patch = v.patch + 1
	}
	vNext.original = v.originalVPrefix() + "" + vNext.String()
	return *vNext
}

// IncMinor produces the next minor Class.
// Sets patch to 0.
// Increments minor number.
// Unsets metadata.
// Unsets prerelease status.
func (v *Class) IncMinor() Class {
	vNext := v
	vNext.metadata = ""
	vNext.pre = ""
	vNext.patch = 0
	vNext.minor = v.minor + 1
	vNext.original = v.originalVPrefix() + "" + vNext.String()
	return *vNext
}

// IncMajor produces the next major Class.
// Sets patch to 0.
// Sets minor to 0.
// Increments major number.
// Unsets metadata.
// Unsets prerelease status.
func (v *Class) IncMajor() Class {
	vNext := v
	vNext.metadata = ""
	vNext.pre = ""
	vNext.patch = 0
	vNext.minor = 0
	vNext.major = v.major + 1
	vNext.original = v.originalVPrefix() + "" + vNext.String()
	return *vNext
}

// SetPrerelease defines the prerelease value.
// Value must not include the required 'hyphen' prefix.
func (v *Class) SetPrerelease(prerelease string) (Class, error) {
	vNext := v
	if len(prerelease) > 0 {
		if err := validatePrerelease(prerelease); err != nil {
			return *vNext, err
		}
	}
	vNext.pre = prerelease
	vNext.original = v.originalVPrefix() + "" + vNext.String()
	return *vNext, nil
}

// SetMetadata defines metadata value.
// Value must not include the required 'plus' prefix.
func (v *Class) SetMetadata(metadata string) (Class, error) {
	vNext := v
	if len(metadata) > 0 {
		if err := validateMetadata(metadata); err != nil {
			return *vNext, err
		}
	}
	vNext.metadata = metadata
	vNext.original = v.originalVPrefix() + "" + vNext.String()
	return *vNext, nil
}

// LessThan tests if one Class is less than another one.
func (v *Class) LessThan(o *Class) bool {
	return v.Compare(o) < 0
}

// GreaterThan tests if one Class is greater than another one.
func (v *Class) GreaterThan(o *Class) bool {
	return v.Compare(o) > 0
}

// Equal tests if two Class are equal to each other.
// Note, Class can be equal with different metadata since metadata
// is not considered part of the comparable Class.
func (v *Class) Equal(o *Class) bool {
	return v.Compare(o) == 0
}

// Compare compares this Class to another one. It returns -1, 0, or 1 if
// the Class smaller, equal, or larger than the other Class.
//
// Class are compared by X.Y.Z. Build metadata is ignored. Prerelease is
// lower than the Class without a prerelease. Compare always takes into account
// predeceases. If you want to work with ranges using typical range syntaxes that
// skip predeceases if the range is not looking for them use constraints.
func (v *Class) Compare(o *Class) int {
	// Compare the major, minor, and patch Class for differences. If a
	// difference is found return the comparison.
	if d := compareSegment(v.Major(), o.Major()); d != 0 {
		return d
	}
	if d := compareSegment(v.Minor(), o.Minor()); d != 0 {
		return d
	}
	if d := compareSegment(v.Patch(), o.Patch()); d != 0 {
		return d
	}

	// At this point the major, minor, and patch Class are the same.
	ps := v.pre
	po := o.Prerelease()

	if ps == "" && po == "" {
		return 0
	}
	if ps == "" {
		return 1
	}
	if po == "" {
		return -1
	}

	return comparePrerelease(ps, po)
}

// UnmarshalJSON implements JSON.Unmarshaler interface.
func (v *Class) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	temp, err := New(s)
	if err != nil {
		return err
	}
	v.major = temp.major
	v.minor = temp.minor
	v.patch = temp.patch
	v.pre = temp.pre
	v.metadata = temp.metadata
	v.original = temp.original
	return nil
}

// MarshalJSON implements JSON.Marshaller interface.
func (v *Class) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (v *Class) UnmarshalText(text []byte) error {
	temp, err := New(string(text))
	if err != nil {
		return err
	}

	*v = *temp

	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (v *Class) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// Scan implements the SQL.Scanner interface.
func (v *Class) Scan(value interface{}) error {
	var s string
	s, _ = value.(string)
	temp, err := New(s)
	if err != nil {
		return err
	}
	v.major = temp.major
	v.minor = temp.minor
	v.patch = temp.patch
	v.pre = temp.pre
	v.metadata = temp.metadata
	v.original = temp.original
	return nil
}

// Value implements the Driver.Valuer interface.
func (v *Class) Value() (driver.Value, error) {
	return v.String(), nil
}

func compareSegment(v, o uint64) int {
	if v < o {
		return -1
	}
	if v > o {
		return 1
	}

	return 0
}

func comparePrerelease(v, o string) int {
	// split the prelease Class by their part. The separator, per the spec,
	// is a .
	sParts := strings.Split(v, ".")
	oPArts := strings.Split(o, ".")

	// Find the longer length of the parts to know how many loop iterations to
	// go through.
	sLen := len(sParts)
	oLen := len(oPArts)

	l := sLen
	if oLen > sLen {
		l = oLen
	}

	// Iterate over each part of the pre-releases to compare the differences.
	for i := 0; i < l; i++ {
		// Since the length of the parts can be different we need to create
		// a placeholder. This is to avoid out of bounds issues.
		sTemp := ""
		if i < sLen {
			sTemp = sParts[i]
		}

		otemp := ""
		if i < oLen {
			otemp = oPArts[i]
		}

		d := comparePrePart(sTemp, otemp)
		if d != 0 {
			return d
		}
	}

	// Reaching here means two Class are of equal value but have different
	// metadata (the part following a +). They are not identical in string form
	// but the Class comparison finds them to be equal.
	return 0
}

func comparePrePart(s, o string) int {
	// Fast path if they are equal
	if s == o {
		return 0
	}

	// When s or o are empty we can use the other in an attempt to determine
	// the response.
	if s == "" {
		if o != "" {
			return -1
		}
		return 1
	}

	if o == "" {
		if s != "" {
			return 1
		}
		return -1
	}

	// When comparing strings "99" is greater than "103". To handle
	// cases like this we need to detect numbers and compare them. According
	// to the semver spec, numbers are always positive. If there is a - at the
	// start like -99 this is to be evaluated as an alphanum. numbers always
	// have precedence over alphanum. Parsing as Uints because negative numbers
	// are ignored.

	oi, n1 := strconv.ParseUint(o, 10, 64)
	si, n2 := strconv.ParseUint(s, 10, 64)

	// The case where both are strings compare the strings
	if n1 != nil && n2 != nil {
		if s > o {
			return 1
		}
		return -1
	} else if n1 != nil {
		// o is a string and s is a number
		return -1
	} else if n2 != nil {
		// s is a string and o is a number
		return 1
	}
	// Both are numbers
	if si > oi {
		return 1
	}
	return -1
}

// Like strings.ContainsAny but does an only instead of any.
func containsOnly(s string, comp string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(comp, r)
	}) == -1
}

// From the spec, "Identifiers MUST comprise only
// ASCII alphanumerics and hyphen [0-9A-Za-z-]. Identifiers MUST NOT be empty.
// Numeric identifiers MUST NOT include leading zeroes.". These segments can
// be dot separated.
func validatePrerelease(p string) error {
	eParts := strings.Split(p, ".")
	for _, p := range eParts {
		if containsOnly(p, num) {
			if len(p) > 1 && p[0] == '0' {
				return ErrSegmentStartsZero
			}
		} else if !containsOnly(p, allowed) {
			return ErrInvalidPrerelease
		}
	}

	return nil
}

// From the spec, "Build metadata MAY be denoted by
// appending a plus sign and a series of dot separated identifiers immediately
// following the patch or pre-release Class. Identifiers MUST comprise only
// ASCII alphanumerics and hyphen [0-9A-Za-z-]. Identifiers MUST NOT be empty."
func validateMetadata(m string) error {
	eParts := strings.Split(m, ".")
	for _, p := range eParts {
		if !containsOnly(p, allowed) {
			return ErrInvalidMetadata
		}
	}
	return nil
}
