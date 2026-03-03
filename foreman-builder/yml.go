package foremanbuilder

import (
	"bytes"
	_ "embed"
	"io"
	"log"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed templates/orbstack-foreman.yml.tmpl
var containerConfigTmpl string

type OrbstackConfigData struct {
	Username      string
	Packages      []string
	InstallString string
}

type Config struct {
	Packages []string `yaml:"packages"`
}

func ParseConfig(r io.Reader) (Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func GetYmlValues(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	return ParseConfig(file)
}

func GenerateContainerConfig(data OrbstackConfigData, pathName string) error {
	tmpl, err := template.New("orbstack-foreman").Parse(containerConfigTmpl)
	if err != nil {
		return err
	}

	if len(data.Packages) != 0 {
		installStr := MakeInstallStringFromStruct(data.Packages)
		log.Printf(installStr, "install str")
		data.InstallString = installStr
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}
	return os.WriteFile(pathName, buf.Bytes(), 0644)
}

func MakeInstallStringFromStruct(packages []string) string {
	var str = ""

	for _, v := range packages {
		str = str + v + " "
	}
	return str

}
