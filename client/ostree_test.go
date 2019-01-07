package client

import (
	"os/exec"
	"strings"
	"testing"
)

func TestOSTreeStatus(t *testing.T) {
	simple := strings.TrimSpace(`
* lmp 435b6162c6240ac995421d0417ebfa79cf0f6081d34f9d995a2431a695ded52b.15
    origin refspec: 435b6162c6240ac995421d0417ebfa79cf0f6081d34f9d995a2431a695ded52b
  lmp f315bbe0cde9125f91ca3faee238df121fbb0ad20499b11148402ee7f0fb1859.0 (rollback)
    origin refspec: f315bbe0cde9125f91ca3faee238df121fbb0ad20499b11148402ee7f0fb1859
`)
	execCommand = NewMockExec(simple, "", 0)
	defer func() { execCommand = exec.Command }()

	status, err := NewOSTreeStatus()
	if err != nil {
		t.Error(err)
	}
	if status.Active != "435b6162c6240ac995421d0417ebfa79cf0f6081d34f9d995a2431a695ded52b" {
		t.Errorf("Invalid value for active image: %s", status.Active)
	}
	if status.Pending != nil {
		t.Errorf("Pending should be nil not: %s", *status.Pending)
	}
}

func TestOSTreeStatusPending(t *testing.T) {
	simple := strings.TrimSpace(`
  lmp 435b6162c6240ac995421d0417ebfa79cf0f6081d34f9d995a2431a695ded52b.0 (pending)
    origin refspec: 435b6162c6240ac995421d0417ebfa79cf0f6081d34f9d995a2431a695ded52b
*  lmp f315bbe0cde9125f91ca3faee238df121fbb0ad20499b11148402ee7f0fb1859.0
    origin refspec: f315bbe0cde9125f91ca3faee238df121fbb0ad20499b11148402ee7f0fb1859
`)
	execCommand = NewMockExec(simple, "", 0)
	defer func() { execCommand = exec.Command }()

	status, err := NewOSTreeStatus()
	if err != nil {
		t.Error(err)
	}
	if status.Active != "f315bbe0cde9125f91ca3faee238df121fbb0ad20499b11148402ee7f0fb1859" {
		t.Errorf("Invalid value for active image: %s", status.Active)
	}
	if status.Pending == nil {
		t.Error("Pending should not be nil")
	} else if *status.Pending != "435b6162c6240ac995421d0417ebfa79cf0f6081d34f9d995a2431a695ded52b" {
		t.Errorf("Invalid value for pending image: %s", *status.Pending)
	}
}
