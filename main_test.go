package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureStdErr redirects stderr while f is executed and returns the captured
// text.  Similar helpers exist in the e2e tests for stdout.
func captureStdErr(f func()) string {
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func Test_runMain_NoRepo(t *testing.T) {
	// create and switch into an empty temporary directory; there's no .git
	tmp, err := os.MkdirTemp("", "norepo")
	assert.NoError(t, err)
	defer os.RemoveAll(tmp)

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	os.Chdir(tmp)

	stderr := captureStdErr(func() { runMain(Merged, false, false) })
	assert.Contains(t, stderr, "no git repository found")
}
