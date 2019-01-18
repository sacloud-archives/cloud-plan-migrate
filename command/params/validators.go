package params

import (
	"github.com/sacloud/cloud-plan-migrate/command"
)

func validateSakuraID(fieldName string, object interface{}) []error {
	return command.ValidateSakuraID(fieldName, object)
}
