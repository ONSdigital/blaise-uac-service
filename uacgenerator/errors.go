package uacgenerator

import (
	"fmt"
	"strings"
)

type ImportError struct {
	InvalidUacs    []string
	InstrumentUacs []string
}

func (importError *ImportError) Error() string {
	var err string
	if len(importError.InvalidUacs) > 0 {
		err = fmt.Sprintf("Cannot import UACs because some were invalid: [%s]", formatSlice(importError.InvalidUacs))
	}
	if len(importError.InstrumentUacs) > 0 {
		if err == "" {
			err = fmt.Sprintf(
				"Cannot import UACs because some were already in use by questionnaires: [%s]",
				formatSlice(importError.InstrumentUacs),
			)
		} else {
			err = fmt.Sprintf(
				"%s and some UACs were already in use by questionnaires: [%s]",
				err, formatSlice(importError.InstrumentUacs),
			)

		}
	}
	return err
}

func (importError *ImportError) HasErrors() bool {
	return len(importError.InvalidUacs) > 0 || len(importError.InstrumentUacs) > 0
}

func formatSlice(input []string) string {
	var quoted []string
	for _, value := range input {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, value))
	}
	return strings.Join(quoted, ", ")
}
