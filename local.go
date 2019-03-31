package tfstate

import (
	"fmt"

	tfState "github.com/hashicorp/terraform/state"
	"github.com/hashicorp/terraform/terraform"
)

// Local represents a local terraform state
type Local struct {
	Path        string
	PersistPath string
	backend     *tfState.LocalState
	state       *terraform.State
}

// Read retrieves and refreshes the remote state and returns terraform.State
func (l *Local) Read(workspace ...string) (*terraform.State, error) {
	if len(workspace) != 0 {
		fmt.Println("[WARN] local backend doesn't support passing in a workspace as workspaces are are created in separate files")
	}

	l.backend = &tfState.LocalState{
		Path: l.Path,
	}

	if l.PersistPath != "" {
		l.backend.PathOut = l.PersistPath
	}

	err := l.backend.RefreshState()
	if err != nil {
		return nil, err
	}

	l.state = l.backend.State()
	return l.state, nil
}

// Write will just perform a persist
func (l *Local) Write(ts *terraform.State) error {
	err := l.backend.WriteState(ts)
	if err != nil {
		return err
	}
	return nil
}
