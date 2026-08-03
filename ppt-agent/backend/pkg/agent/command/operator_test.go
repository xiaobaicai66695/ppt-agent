package command

import (
	"reflect"
	"testing"
)

func TestLinuxShellCommand(t *testing.T) {
	want := []string{"/bin/sh", "-c", "python3 script.py --check"}
	if got := linuxShellCommand([]string{"python3", "script.py", "--check"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("linuxShellCommand() = %#v, want %#v", got, want)
	}
}
