package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func Test_DeletingBranchesWhenTheCheckOptionIsFalse(t *testing.T) {
	onlyCI(t)

	actual := captureOutput(func() { runMain(false) })

	time.Sleep(5 * time.Second)

	expected := fmt.Sprintf("%s %s", green("✔"), "Deleting branches...")
	assert.Contains(t, actual, expected)
}

func Test_DoNotDeleteBranchesWhenTheCheckOptionIsTrue(t *testing.T) {
	onlyCI(t)

	actual := captureOutput(func() { runMain(true) })

	time.Sleep(5 * time.Second)

	expected := fmt.Sprintf("%s %s", hiBlack("-"), "Deleting branches...")
	assert.Contains(t, actual, expected)
}

func onlyCI(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("skipping test in local")
	}
}

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w

	f()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
