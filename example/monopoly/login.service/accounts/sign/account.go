package sign

type Account struct {
	UID         string
	DisplayName string
	Externs     map[string]string // key/value
}
