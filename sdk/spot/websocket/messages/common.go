package messages

// Struct used to parse sequence numbers in private messages
type SequenceId struct {
	Sequence int64 `json:"sequence"`
}
