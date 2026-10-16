package web

import (
	"errors"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// значения атрибутов type, inputmode и autocomplete у полей форм
const (
	inputText        string = "text"
	inputTel         string = "tel"
	inputNumber      string = "number"
	inputDateTime    string = "datetime-local"
	inputModeNumeric string = "numeric"
	autocompleteOff  string = "off"
)

// field - поле формы для шаблона "field"
type field struct {
	Name         string
	Label        string
	Type         string
	Value        string
	Placeholder  string
	Autocomplete string
	InputMode    string
	Min          string
	Hint         string
	Required     bool
	Error        string
}

// fieldErrors - превращает ошибки ozzo-validation в map "имя поля в структуре" -> текст ошибки
func fieldErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	var errs validation.Errors
	if !errors.As(err, &errs) {
		return map[string]string{"": err.Error()}
	}

	res := make(map[string]string, len(errs))
	for name, e := range errs {
		res[name] = e.Error()
	}

	return res
}

// intAtLeast - правило: строка это целое число не меньше minValue
func intAtLeast(minValue int64, msg string) validation.Rule {
	return validation.By(func(value any) error {
		s, _ := value.(string)
		if s == "" {
			return nil
		}

		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil || n < minValue {
			return errors.New(msg)
		}

		return nil
	})
}
