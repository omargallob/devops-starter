package cli

import (
	"bytes"
	"testing"
)

// executeCommand constructs a root command, sets the given args, captures
// stdout/stderr, and returns them along with any error from Execute.
func executeCommand(t *testing.T, args ...string) (stdout string, stderr string, err error) {
	t.Helper()

	root := NewRootCmd()
	root.SetArgs(args)

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)

	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}
