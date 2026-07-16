package domain

type ErrorKind string

const (
	ErrorValidation ErrorKind = "validation"
	ErrorConflict   ErrorKind = "conflict"
	ErrorForbidden  ErrorKind = "forbidden"
	ErrorNotFound   ErrorKind = "not_found"
)

type Error struct {
	Kind    ErrorKind
	Message string
}

func (err Error) Error() string {
	return err.Message
}
