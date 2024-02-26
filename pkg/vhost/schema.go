package vhost

import (
	"errors"
	"fmt"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

type InputSchema map[string]Definition

type Definition struct {
	Pattern         string `yaml:"pattern"`
	CustomPattern   string `yaml:"custom_pattern"`
	Description     string `yaml:"description"`
	Value           string `yaml:"value"`
	ProvisionerOnly bool   `yaml:"provisioner_only"`
}

var Patterns = map[string]string{
	"domain":             `^[a-zA-Z0-9-_]+(\.[a-zA-Z0-9-_]+)*$`,
	"version":            `^[0-9]+(\.[0-9]+)*$`,
	"endpoint":           `^(?:\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}|[a-zA-Z0-9-_]+(\.[a-zA-Z0-9-_]+)*)(?::\d{1,5})$`,
	"yes-no":             `^(yes|no)$`,
	"ipv4":               `^(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.?\b){4}$`,
	"ipv4-cidr":          `^(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.?\b){4}/(?:3[0-2]|[1-2][0-9]|[0-9])$`,
	"ipv4-cidr-optional": `^(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.?\b){4}(?:/(?:3[0-2]|[1-2][0-9]|[0-9]))?$`,
}

var (
	ErrMissingPattern       = errors.New(`must specify "pattern" or "custom_pattern"`)
	ErrMultiplePatterns     = errors.New(`must specify either "pattern" or "custom_pattern", not both`)
	ErrInvalidPattern       = errors.New(`invalid pattern`)
	ErrInvalidCustomPattern = errors.New(`invalid custom pattern`)
	ErrReservedKey          = errors.New("site_name is a reserved schema key name")
)

func ValidateDefinition(definition Definition) error {
	if definition.Pattern == "" && definition.CustomPattern == "" {
		return ErrMissingPattern
	}
	if definition.Pattern != "" && definition.CustomPattern != "" {
		return ErrMultiplePatterns
	}
	if definition.Pattern != "" {
		if _, ok := Patterns[definition.Pattern]; !ok {
			return ErrInvalidPattern
		}
	}
	if definition.CustomPattern != "" {
		if _, err := regexp.Compile(definition.CustomPattern); err != nil {
			return ErrInvalidCustomPattern
		}
	}
	return nil
}

func ParseInputSchema(input []byte) (InputSchema, error) {
	var schema InputSchema
	err := yaml.Unmarshal(input, &schema)
	if err != nil {
		return schema, err
	}
	for key, definition := range schema {
		if key == "site_name" {
			return schema, ErrReservedKey
		}
		if err := ValidateDefinition(definition); err != nil {
			return schema, errors.New(key + ": " + err.Error())
		}
	}
	return schema, nil
}

func (schema InputSchema) SortedKeys() []string {
	keys := []string{}
	for key := range schema {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (schema InputSchema) Validate(input TemplateInput) (TemplateInput, error) {
	validated := TemplateInput{}
	for key, definition := range schema {
		pattern := definition.CustomPattern
		if pattern == "" {
			pattern = Patterns[definition.Pattern]
		}
		value := definition.Value
		if _, ok := input[key]; ok {
			value = input[key]
		}
		re := regexp.MustCompile(pattern)
		if !re.MatchString(value) {
			return TemplateInput{}, fmt.Errorf("invalid value for input %q", key)
		}
		validated[key] = value
	}
	return validated, nil
}

func (schema InputSchema) Hash() string {
	data := []byte{}
	for _, key := range schema.SortedKeys() {
		schema := schema[key]
		data = append(data, []byte(fmt.Sprintf("%#v", schema))...)
	}
	return HashData(data)
}

func (schema InputSchema) Yaml() []byte {
	data, _ := yaml.Marshal(schema)
	return data
}
