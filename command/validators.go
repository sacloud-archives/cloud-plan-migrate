package command

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func ValidateInStrValues(fieldName string, object interface{}, allowValues ...string) []error {
	res := []error{}

	// if target is nil , return OK(Use required attr if necessary)
	if object == nil {
		return res
	}

	if v, ok := object.(string); ok {
		if v == "" {
			return res
		}

		exists := false
		for _, allow := range allowValues {
			if v == allow {
				exists = true
				break
			}
		}
		if !exists {
			err := fmt.Errorf("%q: must be in [%s]", fieldName, strings.Join(allowValues, ","))
			res = append(res, err)
		}
	}
	return res
}

func ValidateRequired(fieldName string, object interface{}) []error {
	if IsEmpty(object) {
		return []error{fmt.Errorf("%q: is required", fieldName)}
	}
	return []error{}
}

func ValidateSakuraID(fieldName string, object interface{}) []error {
	res := []error{}
	idLen := 12

	// if target is nil , return OK(Use required attr if necessary)
	if object == nil {
		return res
	}

	if id, ok := object.(int64); ok {
		if id == 0 {
			return res
		}
		s := fmt.Sprintf("%d", id)
		strlen := utf8.RuneCountInString(s)
		if id < 0 || strlen != idLen {
			res = append(res, fmt.Errorf("%q: Resource ID must be a %d digits number", fieldName, idLen))
		}
	}

	return res
}
