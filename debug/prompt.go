package debug

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

type SelectWidget = promptui.Select

type SelectFn func(*SelectWidget)

// Select from menu
func Select(label string, choices []string, opt ...SelectFn) (int, string) {
	prompt := promptui.Select{
		Label: label,
		Items: choices,
	}
	for _, fn := range opt {
		fn(&prompt)
	}

	_, result, _ := prompt.Run()

	for i, v := range choices {
		if v == result {
			return i, v
		}
	}
	return -1, ""
}

// Confirm with y/n
func Confirm(label string, defaultY bool) bool {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	if defaultY {
		prompt.Default = "y"
	} else {
		prompt.Default = "n"
	}

	result, _ := prompt.Run()

	result = strings.ToLower(result)
	if defaultY {
		return result != "n"
	}
	return !(result != "y")
}

// InputPassword with mask
func InputPassword(label string, validateFunc func(string) error) string {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validateFunc,
		Mask:     '*',
	}

	result, err := prompt.Run()

	if err != nil {
		panic(fmt.Sprintf("When input password %s:%v", label, err))
	}

	return result
}

// Input text
func Input(label string, validateFunc func(string) error) string {
	prompt := promptui.Prompt{
		Label:    label,
		Validate: validateFunc,
	}

	result, err := prompt.Run()

	if err != nil {
		panic(fmt.Sprintf("When input password %s:%v", label, err))
	}

	return result
}
