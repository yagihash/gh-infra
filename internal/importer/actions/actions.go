package actions

// Action is the user-facing write-back choice in the interactive viewer.
type Action string

const (
	Write Action = "write"
	Patch Action = "patch"
	Skip  Action = "skip"
)
