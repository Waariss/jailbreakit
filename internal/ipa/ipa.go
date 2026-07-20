package ipa

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
)

type Report struct {
	Path                        string
	AppPath                     string
	BundleIdentifier            string
	DisplayName                 string
	ShortVersion                string
	BuildVersion                string
	MinimumOSVersion            string
	Executable                  string
	Architectures               []string
	Entitlements                []string
	EmbeddedProvisioningProfile bool
	Warnings                    []string
}

func Inspect(ipaPath string) (Report, error) {
	if strings.TrimSpace(ipaPath) == "" {
		return Report{}, fmt.Errorf("missing IPA path")
	}
	if !strings.HasSuffix(strings.ToLower(ipaPath), ".ipa") {
		return Report{}, fmt.Errorf("IPA path must end with .ipa")
	}
	archive, err := zip.OpenReader(ipaPath)
	if err != nil {
		return Report{}, err
	}
	defer archive.Close()

	infoEntry := findInfoPlist(archive.File)
	if infoEntry == nil {
		return Report{}, fmt.Errorf("IPA does not contain Payload/<app>.app/Info.plist")
	}
	infoData, err := readEntry(infoEntry)
	if err != nil {
		return Report{}, fmt.Errorf("read Info.plist: %w", err)
	}
	values, err := parsePlist(infoData)
	if err != nil {
		return Report{}, fmt.Errorf("parse Info.plist: %w", err)
	}

	appPath := strings.TrimSuffix(infoEntry.Name, "/Info.plist")
	report := Report{
		Path:             ipaPath,
		AppPath:          appPath,
		BundleIdentifier: stringValue(values["CFBundleIdentifier"]),
		DisplayName:      firstString(values["CFBundleDisplayName"], values["CFBundleName"]),
		ShortVersion:     stringValue(values["CFBundleShortVersionString"]),
		BuildVersion:     stringValue(values["CFBundleVersion"]),
		MinimumOSVersion: stringValue(values["MinimumOSVersion"]),
		Executable:       stringValue(values["CFBundleExecutable"]),
	}
	if report.Executable == "" {
		report.Warnings = append(report.Warnings, "CFBundleExecutable is missing from Info.plist")
	} else if executable := findEntry(archive.File, path.Join(appPath, report.Executable)); executable != nil {
		data, readErr := readEntry(executable)
		if readErr == nil {
			report.Architectures = architectures(data)
		}
	}
	if len(report.Architectures) == 0 {
		report.Warnings = append(report.Warnings, "could not determine app architectures from the embedded executable")
	}

	provisioning := findEntry(archive.File, path.Join(appPath, "embedded.mobileprovision"))
	if provisioning != nil {
		report.EmbeddedProvisioningProfile = true
		if data, readErr := readEntry(provisioning); readErr == nil {
			if profile, parseErr := parseEmbeddedProfile(data); parseErr == nil {
				if entitlements, ok := profile["Entitlements"].(map[string]any); ok {
					report.Entitlements = mapKeys(entitlements)
				}
			}
		}
	}
	return report, nil
}

func Print(w io.Writer, report Report) {
	fmt.Fprintf(w, "IPA:                         %s\n", report.Path)
	fmt.Fprintf(w, "App:                         %s\n", report.AppPath)
	fmt.Fprintf(w, "Bundle ID:                   %s\n", valueOrUnknown(report.BundleIdentifier))
	fmt.Fprintf(w, "Display name:                %s\n", valueOrUnknown(report.DisplayName))
	fmt.Fprintf(w, "Version:                     %s\n", versionLine(report.ShortVersion, report.BuildVersion))
	fmt.Fprintf(w, "Minimum iOS:                 %s\n", valueOrUnknown(report.MinimumOSVersion))
	fmt.Fprintf(w, "Executable:                  %s\n", valueOrUnknown(report.Executable))
	fmt.Fprintf(w, "Architectures:               %s\n", joinOrUnknown(report.Architectures))
	if report.EmbeddedProvisioningProfile {
		fmt.Fprintln(w, "Embedded provisioning profile: present")
	} else {
		fmt.Fprintln(w, "Embedded provisioning profile: not found")
	}
	fmt.Fprintf(w, "Entitlements:                %s\n", joinOrUnknown(report.Entitlements))
	for _, warning := range report.Warnings {
		fmt.Fprintf(w, "[!] %s\n", warning)
	}
	fmt.Fprintln(w, "[*] Inspection only: the IPA was not modified or installed.")
}

func findInfoPlist(files []*zip.File) *zip.File {
	var matches []*zip.File
	for _, file := range files {
		if strings.HasPrefix(file.Name, "Payload/") && strings.HasSuffix(file.Name, ".app/Info.plist") {
			matches = append(matches, file)
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Name < matches[j].Name })
	if len(matches) == 0 {
		return nil
	}
	return matches[0]
}

func findEntry(files []*zip.File, name string) *zip.File {
	for _, file := range files {
		if file.Name == name {
			return file
		}
	}
	return nil
}

func readEntry(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func parsePlist(data []byte) (map[string]any, error) {
	if bytes.HasPrefix(data, []byte("bplist00")) {
		value, err := parseBinaryPlist(data)
		if err != nil {
			return nil, err
		}
		result, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("binary plist root is not a dictionary")
		}
		return result, nil
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if start, ok := token.(xml.StartElement); ok && start.Name.Local == "dict" {
			value, err := parseValue(decoder, start)
			if err != nil {
				return nil, err
			}
			result, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("plist root is not a dictionary")
			}
			return result, nil
		}
	}
}

func parseValue(decoder *xml.Decoder, start xml.StartElement) (any, error) {
	switch start.Name.Local {
	case "dict":
		result := make(map[string]any)
		var key string
		for {
			token, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			switch value := token.(type) {
			case xml.StartElement:
				if value.Name.Local == "key" {
					if err := decoder.DecodeElement(&key, &value); err != nil {
						return nil, err
					}
					continue
				}
				parsed, err := parseValue(decoder, value)
				if err != nil {
					return nil, err
				}
				if key != "" {
					result[key] = parsed
					key = ""
				}
			case xml.EndElement:
				if value.Name.Local == "dict" {
					return result, nil
				}
			}
		}
	case "array":
		var result []any
		for {
			token, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			switch value := token.(type) {
			case xml.StartElement:
				parsed, err := parseValue(decoder, value)
				if err != nil {
					return nil, err
				}
				result = append(result, parsed)
			case xml.EndElement:
				if value.Name.Local == "array" {
					return result, nil
				}
			}
		}
	case "string", "integer", "real", "date", "data":
		var text string
		if err := decoder.DecodeElement(&text, &start); err != nil {
			return nil, err
		}
		return strings.TrimSpace(text), nil
	case "true", "false":
		return start.Name.Local == "true", skipElement(decoder, start)
	default:
		return nil, skipElement(decoder, start)
	}
}

func skipElement(decoder *xml.Decoder, start xml.StartElement) error {
	var ignored any
	return decoder.DecodeElement(&ignored, &start)
}

func parseEmbeddedProfile(data []byte) (map[string]any, error) {
	start := bytes.Index(data, []byte("<plist"))
	end := bytes.Index(data, []byte("</plist>"))
	if start < 0 || end < start {
		return nil, fmt.Errorf("embedded provisioning profile does not contain an XML plist")
	}
	end += len("</plist>")
	return parsePlist(data[start:end])
}

func architectures(data []byte) []string {
	if len(data) < 8 {
		return nil
	}
	magic := binary.BigEndian.Uint32(data[:4])
	if magic == 0xcafebabe || magic == 0xcafebabf {
		count := binary.BigEndian.Uint32(data[4:8])
		entrySize := uint32(20)
		if magic == 0xcafebabf {
			entrySize = 32
		}
		var result []string
		for i := uint32(0); i < count && int(8+i*entrySize+4) <= len(data); i++ {
			cpu := binary.BigEndian.Uint32(data[8+i*entrySize:])
			result = appendUnique(result, cpuName(cpu))
		}
		return result
	}
	if magic == 0xbebafeca || magic == 0xbfbafeca {
		cpu := binary.LittleEndian.Uint32(data[4:8])
		return []string{cpuName(cpu)}
	}
	if magic == 0xfeedface || magic == 0xfeedfacf {
		return []string{cpuName(binary.BigEndian.Uint32(data[4:8]))}
	}
	if magic == 0xcefaedfe || magic == 0xcffaedfe {
		return []string{cpuName(binary.LittleEndian.Uint32(data[4:8]))}
	}
	return nil
}

func cpuName(cpu uint32) string {
	switch cpu {
	case 7:
		return "i386"
	case 0x01000007:
		return "x86_64"
	case 12:
		return "arm"
	case 0x0100000c:
		return "arm64"
	default:
		return fmt.Sprintf("cputype-0x%x", cpu)
	}
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func mapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func firstString(values ...any) string {
	for _, value := range values {
		if text := stringValue(value); text != "" {
			return text
		}
	}
	return ""
}

func versionLine(short, build string) string {
	if short == "" {
		short = "unknown"
	}
	if build == "" || build == short {
		return short
	}
	return short + " (build " + build + ")"
}

func joinOrUnknown(values []string) string {
	if len(values) == 0 {
		return "unknown"
	}
	return strings.Join(values, ", ")
}

func valueOrUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

func parseBinaryPlist(data []byte) (any, error) {
	if len(data) < 40 || !bytes.HasPrefix(data, []byte("bplist00")) {
		return nil, fmt.Errorf("invalid binary plist")
	}
	trailer := data[len(data)-32:]
	offsetSize := int(trailer[6])
	refSize := int(trailer[7])
	numObjects := readUint64(trailer[8:16])
	topObject := readUint64(trailer[16:24])
	offsetTable := readUint64(trailer[24:32])
	if offsetSize < 1 || refSize < 1 || numObjects == 0 || topObject >= numObjects || offsetTable >= uint64(len(data)) {
		return nil, fmt.Errorf("invalid binary plist trailer")
	}
	offsets := make([]uint64, numObjects)
	for i := uint64(0); i < numObjects; i++ {
		start := offsetTable + i*uint64(offsetSize)
		end := start + uint64(offsetSize)
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("invalid binary plist offset table")
		}
		offsets[i] = readSizedUint(data[start:end])
	}
	return parseBinaryObject(data, offsets, refSize, topObject, map[uint64]bool{})
}

func parseBinaryObject(data []byte, offsets []uint64, refSize int, index uint64, active map[uint64]bool) (any, error) {
	if index >= uint64(len(offsets)) || active[index] {
		return nil, fmt.Errorf("invalid binary plist object reference")
	}
	start := offsets[index]
	if start >= uint64(len(data)) {
		return nil, fmt.Errorf("invalid binary plist object offset")
	}
	active[index] = true
	defer delete(active, index)
	marker := data[start]
	length, header, err := binaryLength(data, start, marker&0x0f)
	if err != nil {
		return nil, err
	}
	if marker>>4 == 0x1 {
		power := marker & 0x0f
		if power > 3 {
			return nil, fmt.Errorf("unsupported binary plist integer size")
		}
		length = uint64(1) << power
		header = 1
	}
	payload := start + uint64(header)
	if payload > uint64(len(data)) {
		return nil, fmt.Errorf("invalid binary plist object payload")
	}
	switch marker >> 4 {
	case 0x0:
		if marker == 0x08 {
			return false, nil
		}
		if marker == 0x09 {
			return true, nil
		}
		return nil, nil
	case 0x4:
		end := payload + length
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("invalid binary plist data")
		}
		return string(data[payload:end]), nil
	case 0x5:
		end := payload + length
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("invalid binary plist ASCII string")
		}
		return string(data[payload:end]), nil
	case 0x6:
		end := payload + length*2
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("invalid binary plist UTF-16 string")
		}
		chars := make([]uint16, length)
		for i := range chars {
			chars[i] = binary.BigEndian.Uint16(data[payload+uint64(i*2):])
		}
		return string(runeString(chars)), nil
	case 0x1:
		end := payload + length
		if end > uint64(len(data)) {
			return nil, fmt.Errorf("invalid binary plist integer")
		}
		return readSizedUint(data[payload:end]), nil
	case 0xa:
		return parseBinaryArray(data, offsets, refSize, payload, length, active)
	case 0xd:
		return parseBinaryDict(data, offsets, refSize, payload, length, active)
	default:
		return nil, fmt.Errorf("unsupported binary plist object type 0x%x", marker>>4)
	}
}

func binaryLength(data []byte, start uint64, nibble byte) (uint64, int, error) {
	if nibble < 0x0f {
		return uint64(nibble), 1, nil
	}
	if start+2 > uint64(len(data)) {
		return 0, 0, fmt.Errorf("invalid binary plist length")
	}
	marker := data[start+1]
	if marker>>4 != 0x1 {
		return 0, 0, fmt.Errorf("invalid binary plist extended length")
	}
	count := uint64(1) << (marker & 0x0f)
	if count > 8 || start+2+count > uint64(len(data)) {
		return 0, 0, fmt.Errorf("unsupported binary plist length")
	}
	return readSizedUint(data[start+2 : start+2+count]), int(2 + count), nil
}

func parseBinaryArray(data []byte, offsets []uint64, refSize int, start, length uint64, active map[uint64]bool) ([]any, error) {
	if length > uint64(len(data)) || start+length*uint64(refSize) > uint64(len(data)) {
		return nil, fmt.Errorf("invalid binary plist array")
	}
	result := make([]any, 0, length)
	for i := uint64(0); i < length; i++ {
		ref := readSizedUint(data[start+i*uint64(refSize) : start+(i+1)*uint64(refSize)])
		value, err := parseBinaryObject(data, offsets, refSize, ref, active)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func parseBinaryDict(data []byte, offsets []uint64, refSize int, start, length uint64, active map[uint64]bool) (map[string]any, error) {
	refsEnd := start + length*uint64(refSize)*2
	if length > uint64(len(data)) || refsEnd > uint64(len(data)) {
		return nil, fmt.Errorf("invalid binary plist dictionary")
	}
	result := make(map[string]any)
	for i := uint64(0); i < length; i++ {
		keyRef := readSizedUint(data[start+i*uint64(refSize) : start+(i+1)*uint64(refSize)])
		valueStart := start + (length+i)*uint64(refSize)
		valueRef := readSizedUint(data[valueStart : valueStart+uint64(refSize)])
		key, err := parseBinaryObject(data, offsets, refSize, keyRef, active)
		if err != nil {
			return nil, err
		}
		value, err := parseBinaryObject(data, offsets, refSize, valueRef, active)
		if err != nil {
			return nil, err
		}
		if keyText, ok := key.(string); ok {
			result[keyText] = value
		}
	}
	return result, nil
}

func readUint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

func readSizedUint(data []byte) uint64 {
	var value uint64
	for _, part := range data {
		value = value<<8 | uint64(part)
	}
	return value
}

func runeString(chars []uint16) []rune {
	result := make([]rune, len(chars))
	for i, char := range chars {
		result[i] = rune(char)
	}
	return result
}
