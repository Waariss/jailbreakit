package device

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Info struct {
	ProductType string
	ModelName   string
	Chip        string
	OSVersion   string
}

var productMap = map[string]Info{
	"iPhone8,1":  {ProductType: "iPhone8,1", ModelName: "iPhone 6s", Chip: "A9"},
	"iPhone8,2":  {ProductType: "iPhone8,2", ModelName: "iPhone 6s Plus", Chip: "A9"},
	"iPhone8,4":  {ProductType: "iPhone8,4", ModelName: "iPhone SE (1st generation)", Chip: "A9"},
	"iPhone9,1":  {ProductType: "iPhone9,1", ModelName: "iPhone 7", Chip: "A10"},
	"iPhone9,2":  {ProductType: "iPhone9,2", ModelName: "iPhone 7 Plus", Chip: "A10"},
	"iPhone9,3":  {ProductType: "iPhone9,3", ModelName: "iPhone 7", Chip: "A10"},
	"iPhone9,4":  {ProductType: "iPhone9,4", ModelName: "iPhone 7 Plus", Chip: "A10"},
	"iPhone10,1": {ProductType: "iPhone10,1", ModelName: "iPhone 8", Chip: "A11"},
	"iPhone10,2": {ProductType: "iPhone10,2", ModelName: "iPhone 8 Plus", Chip: "A11"},
	"iPhone10,3": {ProductType: "iPhone10,3", ModelName: "iPhone X", Chip: "A11"},
	"iPhone10,4": {ProductType: "iPhone10,4", ModelName: "iPhone 8", Chip: "A11"},
	"iPhone10,5": {ProductType: "iPhone10,5", ModelName: "iPhone 8 Plus", Chip: "A11"},
	"iPhone10,6": {ProductType: "iPhone10,6", ModelName: "iPhone X", Chip: "A11"},
}

func Detect() (Info, error) {
	if _, ok := LookPath("ideviceinfo"); ok {
		return detectWithIdeviceinfo()
	}
	if _, ok := LookPath("palera1n"); ok {
		return detectWithPalera1n()
	}
	return Info{}, fmt.Errorf("no detector found; install libimobiledevice or palera1n")
}

func LookPath(name string) (string, bool) {
	path, err := exec.LookPath(name)
	return path, err == nil
}

func detectWithIdeviceinfo() (Info, error) {
	out, err := exec.Command("ideviceinfo").Output()
	if err != nil {
		return Info{}, fmt.Errorf("ideviceinfo failed: %w", err)
	}
	values := parseKeyValue(out, ":")
	info := Enrich(Info{
		ProductType: values["ProductType"],
		OSVersion:   values["ProductVersion"],
	})
	return info, nil
}

func detectWithPalera1n() (Info, error) {
	out, err := exec.Command("palera1n", "-I").CombinedOutput()
	if err != nil {
		return Info{}, fmt.Errorf("palera1n -I failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	values := parseKeyValue(out, ":")
	info := Enrich(Info{
		ProductType: firstNonEmpty(values["ProductType"], values["Identifier"]),
		OSVersion:   firstNonEmpty(values["ProductVersion"], values["Version"], values["iOS"]),
	})
	return info, nil
}

func parseKeyValue(out []byte, separator string) map[string]string {
	values := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		before, after, ok := strings.Cut(line, separator)
		if !ok {
			continue
		}
		values[strings.TrimSpace(before)] = strings.TrimSpace(after)
	}
	return values
}

func Enrich(info Info) Info {
	product, ok := productMap[info.ProductType]
	if !ok {
		return info
	}
	product.OSVersion = info.OSVersion
	return product
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
