package ad

type Group struct {
	Object

	SamAccountName string
	DisplayName    string
	Description    string
	//GroupCategory  string // FIX ME! json returns int not string
	OrgUnit OrgUnit
	Groups  []Group
	Members []string
}
