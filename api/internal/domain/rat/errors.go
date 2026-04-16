package rat

import "errors"

var (
	ErrNotFound       = errors.New("rat meeting not found")
	ErrQuorumNotMet   = errors.New("quorum not met")
	ErrAlreadyVoted   = errors.New("member has already voted")
)
