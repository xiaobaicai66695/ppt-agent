package evaluation

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestVerifyBundleAcceptsDeclaredCase(t *testing.T) {
	caseJSON := []byte(`{"id":"planner_online_001","name":"线上规划","input":{"user_request":"制作季度复盘"}}`)
	digest := sha256.Sum256(caseJSON)
	manifest := fmt.Sprintf(`{"format":1,"dataset_id":"dataset-1","name":"active-dev-2026w40","role":"active-dev","cases":[{"id":"case-1","suite":"planner","path":"cases/case-1.json","sha256":"%s"}]}`, hex.EncodeToString(digest[:]))
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	caseWriter, err := writer.Create("cases/case-1.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := caseWriter.Write(caseJSON); err != nil {
		t.Fatal(err)
	}
	manifestWriter, err := writer.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manifestWriter.Write([]byte(manifest)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	got, files, err := VerifyBundle(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "active-dev-2026w40" || string(files["cases/case-1.json"]) != string(caseJSON) {
		t.Fatalf("bundle = %#v files = %#v", got, files)
	}
}

func TestVerifyBundleRejectsChecksumMismatch(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	caseWriter, _ := writer.Create("cases/case-1.json")
	_, _ = caseWriter.Write([]byte(`{"id":"case"}`))
	manifestWriter, _ := writer.Create("manifest.json")
	_, _ = manifestWriter.Write([]byte(`{"format":1,"dataset_id":"dataset-1","name":"active","role":"active-dev","cases":[{"id":"case-1","suite":"planner","path":"cases/case-1.json","sha256":"bad"}]}`))
	_ = writer.Close()
	if _, _, err := VerifyBundle(bytes.NewReader(buffer.Bytes())); err == nil {
		t.Fatal("expected checksum failure")
	}
}
