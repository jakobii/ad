package ad

import "strings"

type OrgUnit struct {
	Object

	City                            string
	Country                         string
	Description                     string
	DisplayName                     string
	PostalCode                      string
	ProtectedFromAccidentalDeletion bool
	State                           string
	StreetAddress                   string
}

// IsRoot returns true if there are no other parent OrgUnit's
func (o OrgUnit) IsRoot() bool {
	if reDC.MatchString(o.DistinguishedName) {
		return true
	}
	return false
}

// Parent returns the parent OrgUnit
func (o *OrgUnit) Parent() OrgUnit {
	pieces := strings.Split(o.DistinguishedName, ",")
	return OrgUnit{
		Object: Object{
			Name:              pieces[0],
			DistinguishedName: strings.Join(pieces[1:], ","),
			Connection:        o.Connection,
		},
	}
}
