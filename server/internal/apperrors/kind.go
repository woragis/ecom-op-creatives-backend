package apperrors

type Kind int

const (
	KindInvalid Kind = iota
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
	KindTooManyRequests
	KindInternal
	KindUnavailable
)
