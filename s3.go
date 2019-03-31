package tfstate

import (
	"fmt"

	tfBackend "github.com/hashicorp/terraform/backend"
	tfRemoteStateS3 "github.com/hashicorp/terraform/backend/remote-state/s3"
	tfConfig "github.com/hashicorp/terraform/config"
	tfState "github.com/hashicorp/terraform/state"
	"github.com/hashicorp/terraform/terraform"
)

// Example usage:
// func main() {
// 	remoteState := &S3{
// 		Region:        "us-west-2",
// 		Bucket:        "drinks-terraform-remote-state-test",
// 		Key:           "prod.tfstate",
// 		DynamoDBTable: "terraform-state-lock-test",
// 		Encrypt:       true,
// 	}
//
// 	state, err := remoteState.Read()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	err = remoteState.Write(state)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	err = remoteState.Persist()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// S3 represents terraform state that's persisted in AWS S3
type S3 struct {
	Region        string `map:"region"`
	Bucket        string `map:"bucket"`
	Key           string `map:"key"`
	DynamoDBTable string `map:"dynamodb_table"`
	Encrypt       bool   `map:"encrypt"`

	AccessKey string `map:"access_key"`
	SecretKey string `map:"secret_key"`
	RoleArn   string `map:"role_arn"`

	backend tfBackend.Backend
	state   tfState.State
	lockID  string
}

// Read retrieves and refreshes the remote state and returns terraform.State
// Supports 0 or 1 workspace names, if multiple workspaces are passed in, Read()
// will take the first one and discard the remaining
func (s *S3) Read(workspace ...string) (*terraform.State, error) {
	var ws string
	var lenWs = len(workspace)

	switch {
	case lenWs == 0:
		ws = "default"
	case lenWs == 1:
		ws = workspace[0]
	case lenWs > 1:
		fmt.Println("[WARN] only supports 1 workspace name, using the first one")
		ws = workspace[0]
	}

	conf, err := StructToMap(s)
	if err != nil {
		return nil, err
	}
	raw, err := tfConfig.NewRawConfig(conf)
	if err != nil {
		return nil, err
	}

	resource := terraform.NewResourceConfig(raw)
	s.backend = tfRemoteStateS3.New()
	err = s.backend.Configure(resource)
	if err != nil {
		return nil, err
	}

	s.state, err = s.backend.State(ws)
	if err != nil {
		return nil, err
	}

	err = s.state.RefreshState()
	if err != nil {
		return nil, err
	}

	state := s.state.State()
	return state, nil
}

// Write saves a state in memory
func (s *S3) Write(ts *terraform.State) error {
	err := s.state.WriteState(ts)
	if err != nil {
		return err
	}
	return nil
}

// Persist saves the state to the remote backend, locking an unlocking before
// and after the persist operation
func (s *S3) Persist() error {
	var err error
	// Lock the remote backend prior to persisting
	lockInfo := tfState.NewLockInfo()
	s.lockID, err = s.state.Lock(lockInfo)
	if err != nil {
		return err
	}

	// Perform the persist operation
	err = s.state.PersistState()
	if err != nil {
		return err
	}

	// Unlock after persisting
	err = s.state.Unlock(s.lockID)
	if err != nil {
		return err
	}

	return err
}
