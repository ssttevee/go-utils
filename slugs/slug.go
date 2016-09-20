package slugs

import (
	"regexp"
	"strings"
	"log"
)

var (
	symbolsRegex *regexp.Regexp
	hyphensRegex *regexp.Regexp
	camelRegex   *regexp.Regexp
)

func init() {
	var err error

	symbolsRegex, err = regexp.Compile("(\\W|[_])+")
	if err != nil {
		log.Fatal(err)
	}

	hyphensRegex, err = regexp.Compile("[-]+")
	if err != nil {
		log.Fatal(err)
	}

	camelRegex, err = regexp.Compile("([A-Z]{2}|[a-z][A-Z]|\\d\\D|\\D\\d)")
	if err != nil {
		log.Fatal(err)
	}
}

func Make(str string) string {
	slug := str

	// replace underscores and non-word chars with hyphens
	slug = symbolsRegex.ReplaceAllString(slug, "-")

	// get rid of consecutive hyphens
	slug = hyphensRegex.ReplaceAllString(slug, "-")

	// add hyphens to camel cased words
	slug = camelRegex.ReplaceAllStringFunc(slug, func(str string) string {
		b := []byte(str)
		return string([]byte{b[0], byte('-'), b[1]})
	});

	// return the result in all lower case
	return strings.Trim(strings.ToLower(slug), "-")
}
