package data

import (
	"testing"
)

func Test_CommonValidate(t *testing.T) {

	bg := NewBloodGlucose()

	errors := bg.Common.validate()

	if len(errors) != 0 {
		t.Fatal("Common.Validate: should have errors for new model item")
	}

}
