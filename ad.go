package ad

import (
	"regexp"
	"strings"

	"github.com/jakobii/ps"
)

type Credential = ps.Credential

var reDC = regexp.MustCompile("^DC=.*$")

func ParseDistinguishedName(dn string) (cn string, ou string) {

	if reDC.MatchString(dn) {
		return "", dn
	}

	pieces := strings.Split(dn, ",")
	cn = pieces[0]
	cn = strings.Replace(cn, "CN=", "", 1)
	cn = strings.Replace(cn, "OU=", "", 1)
	return cn, strings.Join(pieces[1:], ",")
}
