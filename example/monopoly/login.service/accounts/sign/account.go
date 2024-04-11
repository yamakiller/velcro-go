package sign

type Account struct {
	UID         string
	DisplayName string
	Rule        int32
	Externs     map[string]string // key/value
}
