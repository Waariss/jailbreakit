package device

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Info struct {
	ProductType string
	ModelName   string
	Chip        string
	OSVersion   string
	Family      string
}

//go:embed product-map.json
var productMapJSON []byte

type productData struct {
	Products map[string]productEntry `json:"products"`
}

type productEntry struct {
	ModelName string `json:"model_name"`
	Chip      string `json:"chip"`
	Family    string `json:"family"`
}

var productMap = loadProductMap()

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

func loadProductMap() map[string]Info {
	var data productData
	if err := json.Unmarshal(productMapJSON, &data); err != nil {
		return map[string]Info{}
	}
	products := make(map[string]Info, len(data.Products))
	for productType, entry := range data.Products {
		products[productType] = Info{
			ProductType: productType,
			ModelName:   entry.ModelName,
			Chip:        entry.Chip,
			Family:      entry.Family,
		}
	}
	return products
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
