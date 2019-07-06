package ad

import (
	"errors"

	"github.com/google/uuid"
)

type Object struct {
	Connection
	Name              string
	ObjectClass       string
	ObjectGuid        uuid.UUID
	DistinguishedName string
}

func (o *Object) Identity() (id string, err error) {

	guid := o.ObjectGuid.String()

	if guid != "00000000-0000-0000-0000-000000000000" {

		return guid, nil

	} else if o.DistinguishedName != "" {

		return o.DistinguishedName, nil

	}

	return "", errors.New("all identity properties are blank")
}

func (o *Object) Pull() error {
	id, err := o.Identity()
	if err != nil {
		return err
	}
	obj, err := o.Connection.GetObject(id)
	if err != nil {
		return err
	}
	o.DistinguishedName = obj.DistinguishedName
	o.Name = obj.Name
	return nil
}
