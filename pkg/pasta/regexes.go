package pasta

import (
	"fmt"
	"regexp"
	"strings"
)

func IncludeRegexp(include string) (*regexp.Regexp, error) {
	include = strings.TrimSpace(include)

	// Make include include everything if unset
	if include == "" {
		include = ".*"
	}

	// Make sure include is embedded in "^..$"
	if include[0] != '^' {
		include = "^" + include
	}

	if include[len(include)-1] != '$' {
		include = include + "$"
	}

	includeRegexp, err := regexp.Compile(include)

	if err != nil {
		return nil, fmt.Errorf("failed to compile include regexp: %w", err)
	}

	return includeRegexp, nil
}

func ExcludeRegexp(exclude string) (*regexp.Regexp, error) {
	exclude = strings.TrimSpace(exclude)

	// Make exclude exclude nothing if unset
	if exclude == "" {
		exclude = "$^"
	} else {
		// Embed exclude embedded in in "^..$", but only if not "exclude nothing"
		if exclude[0] != '^' {
			exclude = "^" + exclude
		}

		if exclude[len(exclude)-1] != '$' {
			exclude = exclude + "$"
		}
	}

	includeRegexp, err := regexp.Compile(exclude)

	if err != nil {
		return nil, fmt.Errorf("failed to compile exclude regexp : %w", err)
	}

	return includeRegexp, nil
}
