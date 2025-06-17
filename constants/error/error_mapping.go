package error

import (
	errField "field-service/constants/error/field"
	errFieldSchedule "field-service/constants/error/field_schedule"
)

func ErrMapping(err error) bool {
	allErrors := make([]error, 0)
	allErrors = append(append(GeneralErrors[:], errField.FieldErrors[:]...), errFieldSchedule.FieldScheduleErrors[:]...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}

	return false
}
