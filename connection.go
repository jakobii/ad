package ad

import (
	"bytes"
	"encoding/json"

	"github.com/jakobii/ps"
)

type Connection struct {
	Server     string
	Credential Credential
}

func NewConnection(Server string, UserName string, Password string) Connection {
	return Connection{
		Server: Server,
		Credential: Credential{
			UserName: UserName,
			Password: Password,
		},
	}
}

func (c *Connection) Test() bool {

	var cmd bytes.Buffer
	cmd.WriteString("Get-ADUser -Server ")
	cmd.WriteString(ps.QuoteString(c.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(c.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(c.Credential.UserName))

	_, err := ps.Invoke(cmd.String())
	if err != nil {
		return false
	}
	return true
}

func (c *Connection) GetObject(Identity string) (obj Object, err error) {

	var cmd bytes.Buffer
	cmd.WriteString("Get-ADObject -Server ")
	cmd.WriteString(ps.QuoteString(c.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(c.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(Identity))
	cmd.WriteString(" | Select-Object @('ObjectGuid', 'ObjectClass',  'DistinguishedName', 'Name') | ConvertTo-Json")

	result, err := ps.Invoke(cmd.String())
	if err != nil {
		return obj, err
	}

	err = json.Unmarshal(result, &obj)
	if err != nil {
		return obj, err
	}

	return obj, nil
}

func (c *Connection) GetUser(Identity string) (user User, err error) {

	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Get-ADUser -Server ")
	cmd.WriteString(ps.QuoteString(c.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(c.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(Identity))
	cmd.WriteString(" -Properties * | Select-Object @('ObjectGuid', 'ObjectClass',  'DistinguishedName', 'Name', 'SamAccountName', 'EmployeeID', 'EmployeeNumber', 'EmailAddress', 'UserPrincipalName', 'AccountExpirationDate', 'Enabled', 'MemberOf', 'CannotChangePassword', 'PasswordNeverExpires', 'PasswordNotRequired', 'DisplayName', 'GivenName', 'Surname', 'OtherName', 'Initials', 'Title', 'Division', 'Department', 'Office', 'Company', 'Organization', 'HomePage', 'Description', 'OfficePhone', 'MobilePhone', 'HomePhone', 'Fax', 'POBox', 'StreetAddress', 'City', 'State', 'PostalCode', 'Country' ) | ConvertTo-Json")

	//fmt.Println(cmd.String())

	result, err := ps.Invoke(cmd.String())
	if err != nil {
		return user, err
	}

	// Object
	err = json.Unmarshal(result, &user.Object)
	if err != nil {
		return user, err
	}

	// User
	err = json.Unmarshal(result, &user)
	if err != nil {
		return user, err
	}

	// OrgUnit
	_, ou := ParseDistinguishedName(user.DistinguishedName)
	user.OrgUnit, err = c.GetOrgUnit(ou)
	if err != nil {
		return user, err
	}

	// []Group
	m := struct{ MemberOf []string }{}
	err = json.Unmarshal(result, &m)
	if err != nil {
		return user, err
	}

	user.Groups = make([]Group, 0, len(m.MemberOf))
	user.originalGroups = make([]Group, 0, len(m.MemberOf))
	for _, v := range m.MemberOf {
		group, err := c.GetGroup(v)
		if err != nil {
			return user, err
		}
		user.Groups = append(user.Groups, group)
		user.originalGroups = append(user.originalGroups, group)
	}

	user.Connection = *c

	return user, nil
}

func (c *Connection) GetOrgUnit(Identity string) (ou OrgUnit, err error) {

	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Get-ADOrganizationalUnit -Server ")
	cmd.WriteString(ps.QuoteString(c.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(c.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(Identity))
	cmd.WriteString(" -Properties * | Select-Object @( 'ObjectGuid', 'ObjectClass',  'DistinguishedName', 'Name', 'City', 'Country', 'Description', 'DisplayName', 'PostalCode', 'ProtectedFromAccidentalDeletion', 'State', 'StreetAddress' ) | ConvertTo-Json")

	//fmt.Println(cmd.String())

	result, err := ps.Invoke(cmd.String())
	if err != nil {
		return ou, err
	}

	// Object
	err = json.Unmarshal(result, &ou.Object)
	if err != nil {
		return ou, err
	}

	// OrgUnit
	err = json.Unmarshal(result, &ou)
	if err != nil {
		return ou, err
	}

	return ou, nil
}

func (c *Connection) GetGroup(Identity string) (group Group, err error) {

	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Get-ADGroup -Server ")
	cmd.WriteString(ps.QuoteString(c.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(c.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(Identity))
	cmd.WriteString(" -Properties * | Select-Object @('ObjectGuid', 'ObjectClass',  'DistinguishedName', 'Name', 'SamAccountName', 'DisplayName', 'Description', 'GroupCategory', 'memberOf', 'members' ) | ConvertTo-Json")

	//fmt.Println(cmd.String())

	result, err := ps.Invoke(cmd.String())
	if err != nil {
		return group, err
	}

	// Object
	err = json.Unmarshal(result, &group.Object)
	if err != nil {
		return group, err
	}

	// Group
	err = json.Unmarshal(result, &group)
	if err != nil {
		return group, err
	}

	// []Group
	m := struct{ MemberOf []string }{}
	err = json.Unmarshal(result, &m)
	if err != nil {
		return group, err
	}
	group.Groups = make([]Group, 0, len(m.MemberOf))
	for _, v := range m.MemberOf {
		group, err := c.GetGroup(v)
		if err != nil {
			return group, err
		}
		group.Groups = append(group.Groups, group)
	}

	return group, nil
}
