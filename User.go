package ad

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jakobii/ps"
)

type User struct {
	Object

	SamAccountName string
	// identity without unique constraint
	EmployeeID        string
	EmployeeNumber    string
	EmailAddress      string
	UserPrincipalName string

	// security
	AccountExpirationDate time.Time
	OrgUnit               OrgUnit
	Enabled               bool
	Groups                []Group
	originalGroups        []Group

	// password
	AccountPassword       string
	ChangePasswordAtLogon bool
	CannotChangePassword  bool
	PasswordNeverExpires  bool
	PasswordNotRequired   bool

	// name
	DisplayName string
	GivenName   string
	Surname     string
	OtherName   string
	Initials    string

	// position
	Title        string
	Division     string
	Department   string
	Office       string
	Company      string
	Organization string
	HomePage     string
	Description  string

	// phone
	OfficePhone string
	MobilePhone string
	HomePhone   string
	Fax         string

	// mail
	POBox         string
	StreetAddress string
	City          string
	State         string
	PostalCode    string
	Country       string
}

func (u *User) Identity() (string, error) {
	guid := u.ObjectGuid.String()

	if guid != "00000000-0000-0000-0000-000000000000" {

		return guid, nil

	} else if u.DistinguishedName != "" {

		return u.DistinguishedName, nil

	} else if u.Name != "" {

		return u.Name, nil

	} else if u.SamAccountName != "" {

		return u.SamAccountName, nil

	}

	return "", errors.New("all identity properties are blank")
}

func (u *User) Pull() error {
	id, err := u.Identity()
	if err != nil {
		return err
	}
	user, err := u.GetUser(id)
	if err != nil {
		return err
	}
	*u = user
	return nil
}

func (u *User) Push() error {
	id, err := u.Identity()
	if err != nil {
		return err
	}

	// password
	if strings.TrimSpace(u.AccountPassword) != "" {
		u.SetPassword(u.AccountPassword)
	}

	// expiration
	u.SetExpiration(u.AccountExpirationDate)

	// leave groups
	var exists bool
	var groupsToRemove []Group
	for _, v := range u.originalGroups {
		vid, err := v.Identity()
		if err != nil {
			return err
		}
		exists = false
		for _, z := range u.Groups {
			zid, err := z.Identity()
			if err != nil {
				return err
			}
			if vid == zid {
				exists = true
			}
		}
		if !exists {
			groupsToRemove = append(groupsToRemove, v)
		}
	}
	err = u.leaveGroups(groupsToRemove)
	if err != nil {
		return err
	}

	// join groups
	var groupsToAdd []Group
	for _, v := range u.Groups {
		vid, err := v.Identity()
		if err != nil {
			return err
		}
		exists = false
		for _, z := range u.originalGroups {
			zid, err := z.Identity()
			if err != nil {
				return err
			}
			if vid == zid {
				exists = true
			}
		}
		if !exists {
			groupsToAdd = append(groupsToAdd, v)
		}
	}
	err = u.joinGroups(groupsToAdd)
	if err != nil {
		return err
	}

	// everything else
	var cmd bytes.Buffer
	cmd.WriteString("Set-ADUser -Server ")
	cmd.WriteString(ps.QuoteString(u.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(u.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(id))
	cmd.WriteString(" -Confirm:$false")

	if u.SamAccountName == "" {
		return errors.New("SamAccountName can not be blank")
	}

	cmd.WriteString(" -SamAccountName ")
	cmd.WriteString(ps.QuoteString(u.SamAccountName))

	cmd.WriteString(" -EmployeeID ")
	cmd.WriteString(ps.QuoteString(u.EmployeeID))

	cmd.WriteString(" -EmployeeNumber ")
	cmd.WriteString(ps.QuoteString(u.EmployeeNumber))

	cmd.WriteString(" -EmailAddress ")
	cmd.WriteString(ps.QuoteString(u.EmailAddress))

	cmd.WriteString(" -UserPrincipalName ")
	cmd.WriteString(ps.QuoteString(u.UserPrincipalName))

	cmd.WriteString(" -Enabled ")
	cmd.WriteString(ps.FormatBool(u.Enabled))

	cmd.WriteString(" -ChangePasswordAtLogon ")
	cmd.WriteString(ps.FormatBool(u.ChangePasswordAtLogon))

	cmd.WriteString(" -CannotChangePassword ")
	cmd.WriteString(ps.FormatBool(u.CannotChangePassword))

	cmd.WriteString(" -PasswordNeverExpires ")
	cmd.WriteString(ps.FormatBool(u.PasswordNeverExpires))

	cmd.WriteString(" -PasswordNotRequired ")
	cmd.WriteString(ps.FormatBool(u.PasswordNotRequired))

	cmd.WriteString(" -DisplayName ")
	cmd.WriteString(ps.QuoteString(u.DisplayName))

	cmd.WriteString(" -GivenName ")
	cmd.WriteString(ps.QuoteString(u.GivenName))

	cmd.WriteString(" -Surname ")
	cmd.WriteString(ps.QuoteString(u.Surname))

	cmd.WriteString(" -OtherName ")
	cmd.WriteString(ps.QuoteString(u.OtherName))

	cmd.WriteString(" -Initials ")
	cmd.WriteString(ps.QuoteString(u.Initials))

	cmd.WriteString(" -Title ")
	cmd.WriteString(ps.QuoteString(u.Title))

	cmd.WriteString(" -Division ")
	cmd.WriteString(ps.QuoteString(u.Division))

	cmd.WriteString(" -Department ")
	cmd.WriteString(ps.QuoteString(u.Department))

	cmd.WriteString(" -Office ")
	cmd.WriteString(ps.QuoteString(u.Office))

	cmd.WriteString(" -Company ")
	cmd.WriteString(ps.QuoteString(u.Company))

	cmd.WriteString(" -Organization ")
	cmd.WriteString(ps.QuoteString(u.Organization))

	cmd.WriteString(" -HomePage ")
	cmd.WriteString(ps.QuoteString(u.HomePage))

	cmd.WriteString(" -Description ")
	cmd.WriteString(ps.QuoteString(u.Description))

	cmd.WriteString(" -OfficePhone ")
	cmd.WriteString(ps.QuoteString(u.OfficePhone))

	cmd.WriteString(" -MobilePhone ")
	cmd.WriteString(ps.QuoteString(u.MobilePhone))

	cmd.WriteString(" -Fax ")
	cmd.WriteString(ps.QuoteString(u.Fax))

	cmd.WriteString(" -POBox ")
	cmd.WriteString(ps.QuoteString(u.POBox))

	cmd.WriteString(" -StreetAddress ")
	cmd.WriteString(ps.QuoteString(u.StreetAddress))

	cmd.WriteString(" -City ")
	cmd.WriteString(ps.QuoteString(u.City))

	cmd.WriteString(" -State ")
	cmd.WriteString(ps.QuoteString(u.State))

	cmd.WriteString(" -PostalCode ")
	cmd.WriteString(ps.QuoteString(u.PostalCode))

	cmd.WriteString(" -Country ")
	cmd.WriteString(ps.QuoteString(u.Country))

	fmt.Println(cmd.String())

	_, err = ps.Invoke(cmd.String())
	if err != nil {
		return err
	}
	return nil
}

func (u *User) SetPassword(NewPassword string) error {
	id, err := u.Identity()
	if err != nil {
		return err
	}
	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Set-ADAccountPassword -Server ")
	cmd.WriteString(ps.QuoteString(u.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(u.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(id))
	cmd.WriteString(" -NewPassword ")
	cmd.WriteString(ps.QuoteString(NewPassword))
	cmd.WriteString(" -Reset -Confirm:$false")

	_, err = ps.Invoke(cmd.String())
	if err != nil {
		return err
	}
	return nil
}

func (u *User) SetExpiration(DateTime time.Time) error {
	if DateTime.IsZero() {
		err := u.ClearExpiration()
		if err != nil {
			return err
		}
		return nil
	}

	id, err := u.Identity()
	if err != nil {
		return err
	}
	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Set-ADAccountExpiration -Server ")
	cmd.WriteString(ps.QuoteString(u.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(u.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(id))
	cmd.WriteString(" -DateTime ")
	cmd.WriteString(ps.QuoteTime(DateTime))
	cmd.WriteString(" -Confirm:$false")

	_, err = ps.Invoke(cmd.String())
	if err != nil {
		return err
	}
	return nil
}

func (u *User) ClearExpiration() error {
	id, err := u.Identity()
	if err != nil {
		return err
	}
	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Clear-ADAccountExpiration -Server ")
	cmd.WriteString(ps.QuoteString(u.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(u.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(id))
	cmd.WriteString(" -Confirm:$false")

	_, err = ps.Invoke(cmd.String())
	if err != nil {
		return err
	}
	return nil
}

func (u *User) joinGroups(groups []Group) error {

	if len(groups) < 1 {
		return nil
	}

	var Members = make([]string, 0, len(groups))
	for _, v := range groups {
		id, err := v.Identity()
		if err != nil {
			return err
		}
		Members = append(Members, ps.QuoteString(id))
	}

	id, err := u.Identity()
	if err != nil {
		return err
	}

	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Add-ADPrincipalGroupMembership -Server ")
	cmd.WriteString(ps.QuoteString(u.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(u.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(id))
	cmd.WriteString(" -MemberOf @(")
	cmd.WriteString(strings.Join(Members, ","))
	cmd.WriteString(")")

	_, err = ps.Invoke(cmd.String())
	if err != nil {
		return err
	}
	return nil
}

func (u *User) leaveGroups(groups []Group) error {

	if len(groups) < 1 {
		return nil
	}

	var Members = make([]string, 0, len(groups))
	for _, v := range groups {
		id, err := v.Identity()
		if err != nil {
			return err
		}
		Members = append(Members, ps.QuoteString(id))
	}

	id, err := u.Identity()
	if err != nil {
		return err
	}

	// get the main stuff
	var cmd bytes.Buffer
	cmd.WriteString("Remove-ADPrincipalGroupMembership -Server ")
	cmd.WriteString(ps.QuoteString(u.Server))
	cmd.WriteString(" -Credential ")
	cmd.WriteString(u.Credential.Expr())
	cmd.WriteString(" -Identity ")
	cmd.WriteString(ps.QuoteString(id))
	cmd.WriteString(" -MemberOf @(")
	cmd.WriteString(strings.Join(Members, ","))
	cmd.WriteString(")")

	_, err = ps.Invoke(cmd.String())
	if err != nil {
		return err
	}
	return nil
}
