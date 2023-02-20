package bizfly

import (
	"regexp"
)

var svgRegexpCompiled *regexp.Regexp

func removeSvgBlock(raw string) (string, error) {
	if svgRegexpCompiled == nil {
		regCompliled, err := regexp.Compile("</svg>:")
		if err != nil {
			return "", err
		}

		svgRegexpCompiled = regCompliled
	}

	loc := svgRegexpCompiled.FindStringIndex(raw)
	if len(loc) == 0 {
		return raw, nil
	}
	return raw[loc[1]:], nil
}

func callBizflyApiWithMeasurement(
	transactionName string,
	callback func() (interface{}, error),
) (interface{}, error) {
	return callback()
}
