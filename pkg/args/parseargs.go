package args

import (
	"io"
	"strings"
	"time"

	kvbuilder "github.com/hashicorp/vault/internalshared/kv-builder"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// these functions borrowed from
// https://github.com/hashicorp/vault/blob/master/command/base_helpers.go

// SanitizePath removes any leading or trailing things from a "path".
func SanitizePath(s string) string {
	return EnsureNoTrailingSlash(EnsureNoLeadingSlash(strings.TrimSpace(s)))
}

// ParseArgsData parses the given args in the format key=value into a map of
// the provided arguments. The given reader can also supply key=value pairs.
func ParseArgsData(stdin io.Reader, args []string) (map[string]interface{}, error) {
	builder := &kvbuilder.Builder{Stdin: stdin}
	if err := builder.Add(args...); err != nil {
		return nil, err
	}

	return builder.Map(), nil
}

// ParseArgsDataString parses the args data and returns the values as strings.
// If the values cannot be represented as strings, an error is returned.
func ParseArgsDataString(stdin io.Reader, args []string) (map[string]string, error) {
	raw, err := ParseArgsData(stdin, args)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := mapstructure.WeakDecode(raw, &result); err != nil {
		return nil, errors.Wrap(err, "failed to convert values to strings")
	}
	if result == nil {
		result = make(map[string]string)
	}
	return result, nil
}

// EnsureTrailingSlash ensures the given string has a trailing slash.
func EnsureTrailingSlash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	for len(s) > 0 && s[len(s)-1] != '/' {
		s = s + "/"
	}
	return s
}

// EnsureNoTrailingSlash ensures the given string has a trailing slash.
func EnsureNoTrailingSlash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// EnsureNoLeadingSlash ensures the given string has a trailing slash.
func EnsureNoLeadingSlash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	for len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	return s
}

// TruncateToSeconds truncates the given duration to the number of seconds. If
// the duration is less than 1s, it is returned as 0. The integer represents
// the whole number unit of seconds for the duration.
func TruncateToSeconds(d time.Duration) int {
	d = d.Truncate(1 * time.Second)

	// Handle the case where someone requested a ridiculously short increment -
	// increments must be larger than a second.
	if d < 1*time.Second {
		return 0
	}

	return int(d.Seconds())
}
