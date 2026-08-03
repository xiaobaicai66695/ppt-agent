package pythonutil

import "testing"

func TestGetPythonBinaryLinuxDefault(t *testing.T) {
	t.Setenv("PYTHON_BIN", "")
	if got := GetPythonBinary(); got != "/root/pptx_env/bin/python" {
		t.Fatalf("GetPythonBinary() = %q", got)
	}
}

func TestGetPythonBinaryOverride(t *testing.T) {
	t.Setenv("PYTHON_BIN", "/opt/ppt/python")
	if got := GetPythonBinary(); got != "/opt/ppt/python" {
		t.Fatalf("GetPythonBinary() = %q", got)
	}
}
