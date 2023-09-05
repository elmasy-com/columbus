/*
fault package is used to predefine common errors in across the columbus services.
*/
package fault

type ColumbusError struct {
	Err string `json:"error"`
}

func (e ColumbusError) Error() string {
	return e.Err
}

var (
	ErrNameEmpty      = ColumbusError{"name is empty"}
	ErrUserNameEmpty  = ColumbusError{"username is empty"}
	ErrDefaultUserNil = ColumbusError{"DefaultUser is nil"}
	ErrUserNil        = ColumbusError{"user is nil"}
	ErrMissingAPIKey  = ColumbusError{"missing API key"}
	ErrInvalidAPIKey  = ColumbusError{"invalid API key"}
	ErrInvalidDomain  = ColumbusError{"invalid domain"}
	ErrPublicSuffix   = ColumbusError{"domain is a public suffix"}
	ErrNotAdmin       = ColumbusError{"not admin"}
	ErrMissingURI     = ColumbusError{"missing URI"}
	ErrBlocked        = ColumbusError{"blocked"}
	ErrNotFound       = ColumbusError{"not found"}
	ErrUserNotFound   = ColumbusError{"user not found"}
	ErrNameTaken      = ColumbusError{"name is taken"}
	ErrBadGateway     = ColumbusError{"bad gateway"}
	ErrGatewayTimeout = ColumbusError{"gateway timeout"}
	ErrUserNotDeleted = ColumbusError{"user not deleted"}
	ErrNotModified    = ColumbusError{"not modified"}
	ErrMultipleUpdate = ColumbusError{"multiple update"}
	ErrSameName       = ColumbusError{"username and name are the same"}
	ErrNothingToDo    = ColumbusError{"nothing to do"}
	ErrConfirmMissing = ColumbusError{"confirmation is missing"}
	ErrNotConfirmed   = ColumbusError{"not confirmed"}
	ErrDataBase       = ColumbusError{"Database error"}
	ErrGetPartsFailed = ColumbusError{"GetParts() failed"}
	ErrInvalidDays    = ColumbusError{"invalid days"}
	ErrTLDOnly        = ColumbusError{"TLD only"}
)
