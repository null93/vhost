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
