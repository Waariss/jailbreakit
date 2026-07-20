package ipa

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectIPA(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Sample.ipa")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	archive := zip.NewWriter(file)
	writeZip(t, archive, "Payload/Sample.app/Info.plist", []byte(`<?xml version="1.0"?><plist><dict>
<key>CFBundleIdentifier</key><string>com.example.sample</string>
<key>CFBundleDisplayName</key><string>Sample</string>
<key>CFBundleShortVersionString</key><string>1.2.3</string>
<key>CFBundleVersion</key><string>42</string>
<key>MinimumOSVersion</key><string>15.0</string>
<key>CFBundleExecutable</key><string>Sample</string>
</dict></plist>`))
	writeZip(t, archive, "Payload/Sample.app/Sample", machoFixture(0x0100000c))
	writeZip(t, archive, "Payload/Sample.app/embedded.mobileprovision", []byte("<plist><dict><key>Entitlements</key><dict><key>application-identifier</key><string>TEAM.com.example.sample</string></dict></dict></plist>"))
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	report, err := Inspect(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.BundleIdentifier != "com.example.sample" || report.ShortVersion != "1.2.3" || report.BuildVersion != "42" {
		t.Fatalf("unexpected metadata: %#v", report)
	}
	if !contains(report.Architectures, "arm64") || !contains(report.Entitlements, "application-identifier") {
		t.Fatalf("missing architecture or entitlement: %#v", report)
	}
	var output bytes.Buffer
	Print(&output, report)
	if !strings.Contains(output.String(), "Inspection only") {
		t.Fatalf("unexpected output: %s", output.String())
	}
}

func TestParseBinaryPlist(t *testing.T) {
	data := binaryPlistFixture("CFBundleIdentifier", "com.example.app")
	values, err := parsePlist(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := stringValue(values["CFBundleIdentifier"]); got != "com.example.app" {
		t.Fatalf("binary plist value = %q", got)
	}
}

func binaryPlistFixture(key, value string) []byte {
	objects := [][]byte{
		{0xd1, 0x01, 0x02},
		binaryString(key),
		binaryString(value),
	}
	result := append([]byte("bplist00"), objects[0]...)
	offsets := []byte{8}
	for _, object := range objects[1:] {
		offsets = append(offsets, byte(len(result)))
		result = append(result, object...)
	}
	offsetTable := len(result)
	result = append(result, offsets...)
	trailer := make([]byte, 32)
	trailer[6] = 1
	trailer[7] = 1
	binary.BigEndian.PutUint64(trailer[8:16], uint64(len(objects)))
	binary.BigEndian.PutUint64(trailer[16:24], 0)
	binary.BigEndian.PutUint64(trailer[24:32], uint64(offsetTable))
	return append(result, trailer...)
}

func binaryString(value string) []byte {
	if len(value) < 15 {
		return append([]byte{0x50 | byte(len(value))}, []byte(value)...)
	}
	return append([]byte{0x5f, 0x10, byte(len(value))}, []byte(value)...)
}

func writeZip(t *testing.T, archive *zip.Writer, name string, content []byte) {
	t.Helper()
	entry, err := archive.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write(content); err != nil {
		t.Fatal(err)
	}
}

func machoFixture(cpu uint32) []byte {
	data := make([]byte, 8)
	data[0] = 0xcf
	data[1] = 0xfa
	data[2] = 0xed
	data[3] = 0xfe
	data[4] = byte(cpu)
	data[5] = byte(cpu >> 8)
	data[6] = byte(cpu >> 16)
	data[7] = byte(cpu >> 24)
	return data
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
