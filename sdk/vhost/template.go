package vhost

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cbroglie/mustache"
	"gopkg.in/yaml.v3"
)

const (
	RE_DOMAIN      = `^[a-zA-Z0-9-_]+(\.[a-zA-Z0-9-_]+)*$`
	RE_PHP_VERSION = `^[0-9]\.[0-9]$`
	RE_ENDPOINT    = `^(?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[a-zA-Z0-9-_]+(\.[a-zA-Z0-9-_]+)*)(?::\d{1,5})$`
	RE_YES_NO      = `^yes|no$`
)

type TemplateSchema map[string]SchemaEntry

func (ts *TemplateSchema) String() string {
	o := ""
	for key, entry := range *ts {
		o += key + ": " + entry.String()
		o += "\n"
	}
	return o
}

func (se *SchemaEntry) String() string {
	value := ""
	pattern := ""
	found := 0
	sep := ""
	if se.Value != nil {
		value = "value: " + *se.Value
		found++
	}
	if se.Pattern != nil {
		value = "pattern: " + *se.Pattern
		found++
	}
	if found > 1 {
		sep = ", "
	}
	return "{ " + value + sep + pattern + " }"
}

type SchemaEntry struct {
	Value   *string
	Pattern *string
	Description *string
}

type template struct {
	Name   string
	Path   string
	Raw    []byte
	Body   string
	Schema TemplateSchema
}

func Validate(schema TemplateSchema, input TemplateInput) (*TemplateInput, error) {
	validated := TemplateInput{}
	for key, entry := range schema {
		_, hasInput := input[key]
		hasDefaultValue := entry.Value != nil
		hasPattern := entry.Pattern != nil
		if !hasDefaultValue && !hasPattern {
			return nil, fmt.Errorf("invalid schema for %q, must have default value or validation pattern", key)
		}
		if hasInput && hasDefaultValue && !hasPattern {
			return nil, fmt.Errorf("unexpected value passed for %q", key)
		}
		if !hasInput && !hasDefaultValue && hasPattern {
			return nil, fmt.Errorf("expecting value for %q", key)
		}
		if !hasInput && hasDefaultValue {
			validated[key] = *entry.Value
			continue
		}

		if strings.HasPrefix(*entry.Pattern, "<") && strings.HasSuffix(*entry.Pattern, ">") {
			switch *entry.Pattern {
			case "<domain>":
				*entry.Pattern = RE_DOMAIN
			case "<php-version>":
				*entry.Pattern = RE_PHP_VERSION
			case "<endpoint>":
				*entry.Pattern = RE_ENDPOINT
			case "<yes-no>":
				*entry.Pattern = RE_YES_NO
			default:
				return nil, fmt.Errorf("unknown pattern %q for input %q", *entry.Pattern, key)
			}
		}
		if strings.HasPrefix(*entry.Pattern, "^") && strings.HasSuffix(*entry.Pattern, "$") {
			if re, err := regexp.Compile(*entry.Pattern); err == nil {
				if !re.MatchString(input[key]) {
					return nil, fmt.Errorf("failed to validate input %q with value %q", key, input[key])
				}
				validated[key] = input[key]
			} else {
				return nil, fmt.Errorf("could not compile pattern for key %q", key)
			}
		} else {
			return nil, fmt.Errorf("invalid pattern for input %q, patterns must start and end with anchors", key)
		}
	}
	return &validated, nil
}

func (t *template) Hash() string {
	hash := sha256.Sum256(t.Raw)
	return fmt.Sprintf("%x", hash[:])
}

func (t *template) EncodeRaw() string {
	return base64.StdEncoding.EncodeToString(t.Raw)
}

func (t *template) Render(domainName string, input TemplateInput) (*VirtualHost, error) {
	if validated, err := Validate(t.Schema, input); err == nil {
		(*validated)["template"] = t.Name
		(*validated)["domain"] = domainName
		vh := VirtualHost{
			Name:     domainName,
			Template: *t,
			Input:    *validated,
		}
		if data, err := mustache.Render(t.Body, *validated); err == nil {
			vh.Body = data
			return &vh, nil
		}
		return nil, fmt.Errorf("failed to template config")
	} else {
		return nil, err
	}
}

func ParseSchema(contents string) (*TemplateSchema, error) {
	schema := TemplateSchema{}
	if err := yaml.Unmarshal([]byte(contents), &schema); err == nil {
		return &schema, nil
	} else {
		return nil, fmt.Errorf("failed parsing template schema")
	}
}

func ParseTemplate(contents []byte) (*template, error) {
	if !strings.HasPrefix(string(contents), "---") {
		return nil, fmt.Errorf("template does not have any metadata attached to it")
	}
	if strings.Count(string(contents), "---") != 2 {
		return nil, fmt.Errorf("template has invalid metadata within it")
	}
	parts := strings.Split(string(contents), "---")
	if schema, err := ParseSchema(strings.TrimSpace(parts[1])); err == nil {
		t := template{
			Raw:    contents,
			Body:   strings.TrimSpace(parts[2]),
			Schema: *schema,
		}
		return &t, nil
	} else {
		return nil, err
	}
}

func ParseEncodedTemplate(encoded string) (*template, error) {
	if decoded, err := base64.StdEncoding.DecodeString(encoded); err == nil {
		return ParseTemplate(decoded)
	}
	return nil, fmt.Errorf("failed to decode base64 encoded template")
}

func LoadTemplate(templateName string) (*template, error) {
	templateName = strings.TrimSuffix (templateName,".tpl")
	templatePath := GetTemplatePath(templateName)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template with name %q was not found on system", templateName)
	}
	if contents, err := os.ReadFile(templatePath); err == nil {
		if t, err := ParseTemplate(contents); err == nil {
			t.Name = templateName
			t.Path = GetTemplatePath(templateName)
			return t, nil
		} else {
			return nil, err
		}
	}
	return nil, fmt.Errorf("could not read template with name %q", templateName)
}
