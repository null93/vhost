package vhost

import (
	"testing"
)

func TestValues(t *testing.T) {
	schema, errValid := ParseInputSchema(
		[]byte(`
            foo:
                pattern: yes-no
                description: foo
                value: foo
                provisioner_only: true
            bar:
                pattern: yes-no
                description: test
                value: maybe
                allowed_values: ["maybe"]
        `),
	)
	if errValid != nil {
		t.Fatal(errValid)
	}
	input := TemplateInput{
		"foo": "yes",
		"bar": "maybe",
	}
	_, errValidate := schema.Validate(input)
	if errValidate != nil {
		t.Fatal(errValidate)
	}
}

func TestParsing(t *testing.T) {
	_, errValid := ParseInputSchema(
		[]byte(`
            foo:
                pattern: yes-no
                description: foo
                value: foo
                provisioner_only: true
            bar:
                custom_pattern: ^bar$
        `),
	)
	if errValid != nil {
		t.Fatal(errValid)
	}
	_, errBlank := ParseInputSchema([]byte(``))
	if errBlank != nil {
		t.Fatal(errBlank)
	}
	_, errReserved := ParseInputSchema([]byte(`site_name: { pattern: "yes-no" }`))
	if errReserved != ErrReservedKey {
		t.Fatal(errReserved)
	}
	_, errAllowedValuesValid := ParseInputSchema([]byte(`foo: { pattern: "yes-no", allowed_values: ["maybe"] }`))
	if errAllowedValuesValid != nil {
		t.Fatal(errAllowedValuesValid)
	}
}

func TestValidateRejectsUnknownInput(t *testing.T) {
	schema, errParse := ParseInputSchema([]byte(`
        domain_names:
          pattern: domain
          value: example.com
    `))
	if errParse != nil {
		t.Fatal(errParse)
	}
	// Typo / mismatched key — schema has domain_names, not domain.
	_, err := schema.Validate(TemplateInput{"domain": "foo.example.com"})
	if err == nil {
		t.Fatal("expected unknown-input error, got nil")
	}
	if want := `unknown input "domain"`; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}

	// site_name is reserved (set by Render) and must always be accepted.
	if _, err := schema.Validate(TemplateInput{"site_name": "anything", "domain_names": "ok.example.com"}); err != nil {
		t.Errorf("site_name should be exempt from the unknown-input check, got: %v", err)
	}
}

func TestValidation(t *testing.T) {
	errMissingPattern := ValidateDefinition(Definition{})
	if errMissingPattern != ErrMissingPattern {
		t.Fatal(errMissingPattern)
	}
	errMultiplePatterns := ValidateDefinition(Definition{
		Pattern:       "yes-no",
		CustomPattern: "^bar$",
	})
	if errMultiplePatterns != ErrMultiplePatterns {
		t.Fatal(errMultiplePatterns)
	}
	errValid := ValidateDefinition(Definition{
		Pattern: "yes-no",
	})
	if errValid != nil {
		t.Fatal(errValid)
	}
	errValid = ValidateDefinition(Definition{
		CustomPattern: "^bar$",
	})
	if errValid != nil {
		t.Fatal(errValid)
	}
	errInvalidPattern := ValidateDefinition(Definition{
		Pattern: "baz",
	})
	if errInvalidPattern != ErrInvalidPattern {
		t.Fatal(errInvalidPattern)
	}
	errValidCustomPattern := ValidateDefinition(Definition{
		CustomPattern: "^bar$",
	})
	if errValidCustomPattern != nil {
		t.Fatal(errValidCustomPattern)
	}
	errInvalidCustomPattern := ValidateDefinition(Definition{
		CustomPattern: "bar(",
	})
	if errInvalidCustomPattern != ErrInvalidCustomPattern {
		t.Fatal(errInvalidCustomPattern)
	}
}
