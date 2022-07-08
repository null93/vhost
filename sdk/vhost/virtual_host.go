package vhost

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type TemplateInput map[string]string

type VirtualHostComment struct {
	TemplateName string
	DomainName   string
	Input        string
	Template     string
}

type VirtualHost struct {
	Name     string
	Template template
	Input    TemplateInput
	Body     string
}

func (ti *TemplateInput) GetEnvironmentalVars() []string {
	vars := []string{}
	for key, value := range *ti {
		re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
		newKey := re.ReplaceAllString(strings.ToUpper(key), "_")
		vars = append(vars, newKey+"="+value)
	}
	return vars
}

func (vh *VirtualHost) Comment() string {
	return fmt.Sprintf(
		"# MANAGED BY jrctl/jetrailsd\n#\n%s\n\n",
		BlockComment(vh.Encode(), 118),
	)
}

func (vh *VirtualHost) EncodeInput() string {
	data, _ := yaml.Marshal(&vh.Input)
	return base64.StdEncoding.EncodeToString(data)
}

func (vh *VirtualHost) Encode() string {
	comment := VirtualHostComment{
		TemplateName: vh.Template.Name,
		DomainName:   vh.Name,
		Input:        vh.EncodeInput(),
		Template:     vh.Template.EncodeRaw(),
	}
	data, _ := yaml.Marshal(&comment)
	var b bytes.Buffer
	writer, _ := flate.NewWriter(&b, flate.BestCompression)
	writer.Write(data)
	writer.Close()
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func (vh *VirtualHost) Save(override bool) error {
	contents := vh.Comment() + vh.Body
	availablePath := GetAvailablePath(vh.Name)
	enabledPath := GetEnabledPath(vh.Name)
	if err := os.WriteFile(availablePath, []byte(contents), 0644); err == nil {
		if err := os.Symlink(availablePath, enabledPath); err != nil && !override {
			return fmt.Errorf("failed enabling virtual host")
		}
		return nil
	}
	return fmt.Errorf("failed saving virtual host to file")
}

func ParseVirtualHost(contents []byte) (*VirtualHost, error) {
	if !strings.HasPrefix(string(contents), "# MANAGED BY jrctl/jetrailsd\n") {
		return nil, fmt.Errorf("template does not have any metadata attached to it")
	}
	encoded := ""
	for _, line := range strings.Split(string(contents), "\n")[1:] {
		if strings.HasPrefix(line, "#") {
			encoded += strings.TrimSpace(strings.TrimPrefix(line, "#"))
		} else {
			break
		}
	}
	if decoded, err := base64.StdEncoding.DecodeString(encoded); err == nil {
		if payload, err := ioutil.ReadAll(flate.NewReader(bytes.NewReader([]byte(decoded)))); err == nil {
			header := VirtualHostComment{}
			if err := yaml.Unmarshal(payload, &header); err == nil {
				if template, err := ParseEncodedTemplate(header.Template); err == nil {
					if decodedInput, err := base64.StdEncoding.DecodeString(header.Input); err == nil {
						input := TemplateInput{}
						if err := yaml.Unmarshal([]byte(decodedInput), &input); err == nil {
							vh := VirtualHost{
								Template: *template,
								Input:    input,
							}
							vh.Name = header.DomainName
							vh.Template.Name = header.TemplateName
							vh.Template.Path = GetTemplatePath(header.TemplateName)
							return &vh, nil
						} else {
							return nil, err
						}
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("failed parsing header metadata")
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func LoadVirtualHost(domainName string) (*VirtualHost, error) {
	availablePath := GetAvailablePath(domainName)
	if _, err := os.Stat(availablePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("virtual host with domain name %q was not found on system", domainName)
	}
	if contents, err := os.ReadFile(availablePath); err == nil {
		return ParseVirtualHost(contents)
	}
	return nil, fmt.Errorf("could not read template with name %q", domainName)
}
