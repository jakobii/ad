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

func (u *User) Push() (err error) {

	// error checking
	//if strings.TrimSpace(u.SamAccountName) == "" {
	//	return errors.New("SamAccountName can not be blank")
	//}
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("Name can not be blank")
	}

	// check if the user object has an object guid
	// not having an object guid is an indication that the
	// user does not yet exist in active directory
	// attempt to create it.
	if u.ObjectGuid.String() == "00000000-0000-0000-0000-000000000000" {
		NameExists, err := u.TestName()
		if err != nil {
			return err
		}
		if NameExists {
			user, err := u.GetUser(u.Name)
			if err != nil {
				return err
			}
			u.Object = user.Object
		} else {
			err = u.NewUser(u.Name)
			if err != nil {
				return err
			}
			user, err := u.GetUser(u.Name)
			if err != nil {
				return err
			}
			u.Object = user.Object
		}
	}

	// update process
	id, err := u.Identity()
	if err != nil {
		return err
	}

	// password
	if strings.TrimSpace(u.AccountPassword) != "" {
		err = u.SetPassword()
		if err != nil {
			return err
		}
	}

	// expiration
	u.SetExpiration(u.AccountExpirationDate)

	// OrgUnit

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
				break
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
				break
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

	//cmd.WriteString(ps.Param("SamAccountName", ps.QuoteString(u.SamAccountName)))
	if u.SamAccountName == "" {
		cmd.WriteString(ps.Param("SamAccountName", "$null"))
	} else {
		cmd.WriteString(ps.Param("SamAccountName", ps.QuoteString(u.SamAccountName)))
	}
	if u.EmployeeID == "" {
		cmd.WriteString(ps.Param("EmployeeID", "$null"))
	} else {
		cmd.WriteString(ps.Param("EmployeeID", ps.QuoteString(u.EmployeeID)))
	}

	if u.EmployeeNumber == "" {
		cmd.WriteString(ps.Param("EmployeeNumber", "$null"))
	} else {
		cmd.WriteString(ps.Param("EmployeeNumber", ps.QuoteString(u.EmployeeNumber)))
	}

	if u.EmailAddress == "" {
		cmd.WriteString(ps.Param("EmailAddress", "$null"))
	} else {
		cmd.WriteString(ps.Param("EmailAddress", ps.QuoteString(u.EmailAddress)))
	}

	if u.UserPrincipalName == "" {
		cmd.WriteString(ps.Param("UserPrincipalName", "$null"))
	} else {
		cmd.WriteString(ps.Param("UserPrincipalName", ps.QuoteString(u.UserPrincipalName)))
	}

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

	if u.DisplayName == "" {
		cmd.WriteString(ps.Param("DisplayName", "$null"))
	} else {
		cmd.WriteString(ps.Param("DisplayName", ps.QuoteString(u.DisplayName)))
	}

	if u.GivenName == "" {
		cmd.WriteString(ps.Param("GivenName", "$null"))
	} else {
		cmd.WriteString(ps.Param("GivenName", ps.QuoteString(u.GivenName)))
	}

	if u.Surname == "" {
		cmd.WriteString(ps.Param("Surname", "$null"))
	} else {
		cmd.WriteString(ps.Param("Surname", ps.QuoteString(u.Surname)))
	}

	if u.OtherName == "" {
		cmd.WriteString(ps.Param("OtherName", "$null"))
	} else {
		cmd.WriteString(ps.Param("OtherName", ps.QuoteString(u.OtherName)))
	}

	if u.Initials == "" {
		cmd.WriteString(ps.Param("Initials", "$null"))
	} else {
		cmd.WriteString(ps.Param("Initials", ps.QuoteString(u.Initials)))
	}

	if u.Title == "" {
		cmd.WriteString(ps.Param("Title", "$null"))
	} else {
		cmd.WriteString(ps.Param("Title", ps.QuoteString(u.Title)))
	}

	if u.Division == "" {
		cmd.WriteString(ps.Param("Division", "$null"))
	} else {
		cmd.WriteString(ps.Param("Division", ps.QuoteString(u.Division)))
	}

	if u.Department == "" {
		cmd.WriteString(ps.Param("Department", "$null"))
	} else {
		cmd.WriteString(ps.Param("Department", ps.QuoteString(u.Department)))
	}

	if u.Office == "" {
		cmd.WriteString(ps.Param("Office", "$null"))
	} else {
		cmd.WriteString(ps.Param("Office", ps.QuoteString(u.Office)))
	}

	if u.Company == "" {
		cmd.WriteString(ps.Param("Company", "$null"))
	} else {
		cmd.WriteString(ps.Param("Company", ps.QuoteString(u.Company)))
	}

	if u.Organization == "" {
		cmd.WriteString(ps.Param("Organization", "$null"))
	} else {
		cmd.WriteString(ps.Param("Organization", ps.QuoteString(u.Organization)))
	}

	if u.HomePage == "" {
		cmd.WriteString(ps.Param("HomePage", "$null"))
	} else {
		cmd.WriteString(ps.Param("HomePage", ps.QuoteString(u.HomePage)))
	}

	if u.Description == "" {
		cmd.WriteString(ps.Param("Description", "$null"))
	} else {
		cmd.WriteString(ps.Param("Description", ps.QuoteString(u.Description)))
	}

	if u.OfficePhone == "" {
		cmd.WriteString(ps.Param("OfficePhone", "$null"))
	} else {
		cmd.WriteString(ps.Param("OfficePhone", ps.QuoteString(u.OfficePhone)))
	}

	if u.MobilePhone == "" {
		cmd.WriteString(ps.Param("MobilePhone", "$null"))
	} else {
		cmd.WriteString(ps.Param("MobilePhone", ps.QuoteString(u.MobilePhone)))
	}

	if u.Fax == "" {
		cmd.WriteString(ps.Param("Fax", "$null"))
	} else {
		cmd.WriteString(ps.Param("Fax", ps.QuoteString(u.Fax)))
	}

	if u.POBox == "" {
		cmd.WriteString(ps.Param("POBox", "$null"))
	} else {
		cmd.WriteString(ps.Param("POBox", ps.QuoteString(u.POBox)))
	}

	if u.StreetAddress == "" {
		cmd.WriteString(ps.Param("StreetAddress", "$null"))
	} else {
		cmd.WriteString(ps.Param("StreetAddress", ps.QuoteString(u.StreetAddress)))
	}

	if u.City == "" {
		cmd.WriteString(ps.Param("City", "$null"))
	} else {
		cmd.WriteString(ps.Param("City", ps.QuoteString(u.City)))
	}

	if u.State == "" {
		cmd.WriteString(ps.Param("State", "$null"))
	} else {
		cmd.WriteString(ps.Param("State", ps.QuoteString(u.State)))
	}

	if u.PostalCode == "" {
		cmd.WriteString(ps.Param("PostalCode", "$null"))
	} else {
		cmd.WriteString(ps.Param("PostalCode", ps.QuoteString(u.PostalCode)))
	}

	if u.Country == "" {
		cmd.WriteString(ps.Param("Country", "$null"))
	} else {
		cmd.WriteString(ps.Param("Country", ps.QuoteString(u.Country)))
	}

	fmt.Println(cmd.String())

	_, err = powershell(cmd.String())
	if err != nil {
		return err
	}
	return nil
}

func (u *User) SetPassword() error {
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
	cmd.WriteString(ps.SecureString(u.AccountPassword))
	cmd.WriteString(" -Reset -Confirm:$false")

	fmt.Println(cmd.String())

	_, err = powershell(cmd.String())
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

	_, err = powershell(cmd.String())
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

	_, err = powershell(cmd.String())
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

	_, err = powershell(cmd.String())
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

	_, err = powershell(cmd.String())
	if err != nil {
		return err
	}
	return nil
}

// TestSamAccountName return true if the SamAccountName exists in active directory
func (u *User) TestSamAccountName() (bool, error) {
	return u.TestADUser("SamAccountName", u.SamAccountName)
}

// TestUserPrincipalName return true if the UserPrincipalName exists in active directory
func (u *User) TestUserPrincipalName() (bool, error) {
	return u.TestADUser("UserPrincipalName", u.UserPrincipalName)
}

// TestEmailAddress return true if the EmailAddress exists in active directory
func (u *User) TestEmailAddress() (bool, error) {
	return u.TestADUser("EmailAddress", u.EmailAddress)
}

// TestName return true if the Name exists in active directory
func (u *User) TestName() (bool, error) {
	return u.TestADUser("Name", u.Name)
}
