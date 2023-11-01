package device

//go:generate stringer -type QueryState -linecomment

type QueryState uint8

const (
	QueryNone    QueryState = iota // none
	QueryStarted                   // started
	QuerySuccess                   // success
	QueryFailed                    // failed
)
