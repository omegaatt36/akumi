package tui

// ViewState represents the different views/screens of the application.
type ViewState int

const (
	// StateListTargets represents the view that shows the list of configured SSH targets.
	StateListTargets ViewState = iota
	// StateCreateTarget represents the view for adding a new target.
	StateCreateTarget
	// StateEditTarget represents the view for editing an existing target.
	StateEditTarget
	// StateConfirmDelete represents the confirmation dialog for deleting a target.
	StateConfirmDelete
)

const (
	// InputUser is the index for the username input field.
	InputUser int = iota
	// InputHost is the index for the hostname input field.
	InputHost
	// InputPort is the index for the port number input field.
	InputPort
	// InputNickname is the index for the optional nickname input field.
	InputNickname
	// NumInputs represents the total number of input fields.
	NumInputs
)
