package debug

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	py "github.com/qjpcpu/common/pinyin"
	"github.com/qjpcpu/go-prompt"
)

const (
	ParamInputHintSymbol = ">"
	PromptTypeFile       = "FILE   "
	PromptTypeDir        = "DIR    "
	PromptTypeDefault    = "DEFAULT"
	PromptTypeHistory    = "HISTORY"
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

// SelectWithSearch from menu
func SelectWithSearch(label string, choices []string) int {
	searchFunction := func(s *SelectWidget) {
		s.Size = 20
		s.IsVimMode = true
		s.HideSelected = true
		s.Searcher = func(input string, index int) bool {
			_, idx := py.FuzzyContain(choices[index], input)
			return idx >= 0
		}
	}
	idx, _ := Select(label, choices, searchFunction)
	return idx
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

	return strings.TrimSpace(result)
}

type InputOption func(*inputOption)

type Suggest struct {
	Text string
	Desc string
}

func (s Suggest) GetKey() string          { return s.Text }
func (s Suggest) convert() prompt.Suggest { return prompt.Suggest{Text: s.Text, Description: s.Desc} }

type inputOption struct {
	recentBucket string
	currentFiles bool
	suggestions  []Suggest
}

func WithRecent(ns string) InputOption {
	return func(opt *inputOption) {
		opt.recentBucket = ns
	}
}

func WithCurrentFiles() InputOption {
	return func(opt *inputOption) {
		opt.currentFiles = true
	}
}

func WithSuggestions(list []Suggest) InputOption {
	return func(opt *inputOption) {
		opt.suggestions = list
	}
}

// Input text
func Input(label string, fns ...InputOption) string {
	opt := new(inputOption)
	for _, fn := range fns {
		fn(opt)
	}
	var cache *DiskCache
	sugMap := make(map[string]Suggest)
	menu := func(d prompt.Document) (ret []prompt.Suggest) {
		suggestions := opt.suggestions
		if opt.currentFiles {
			files, _ := ioutil.ReadDir(".")
			for _, file := range files {
				if strings.HasPrefix(file.Name(), ".") {
					continue
				}
				suggestions = append(suggestions, Suggest{Text: file.Name(), Desc: fileDesc(file)})
			}
		}
		if strings.TrimSpace(opt.recentBucket) != "" {
			opt.recentBucket = strings.TrimSpace(opt.recentBucket)
			cache, _ = NewHomeDiskCache()
			cache.SetBucketSize(opt.recentBucket, 5)
			var history []Suggest
			cache.ListItem(opt.recentBucket, &history)
			dup := make(map[string]*Suggest)
			if len(history) > 0 {
				for i, item := range history {
					dup[item.Text] = &history[i]
				}
				for _, sug := range suggestions {
					if v, ok := dup[sug.Text]; !ok {
						history = append(history, sug)
					} else {
						v.Desc = sug.Desc
					}
				}
				suggestions = history
			}
		}
		for _, s := range suggestions {
			ret = append(ret, s.convert())
			sugMap[s.Text] = s
		}
		return
	}
	text, _ := prompt.Input(
		label+" ",
		menu,
		prompt.OptionPrefixTextColor(prompt.Blue),
	)
	text = strings.TrimSpace(text)
	if cache != nil {
		defer cache.Close()
		if text != "" {
			sug := sugMap[text]
			if sug.Text == "" {
				sug.Text = text
				sug.Desc = PromptTypeHistory
			}
			cache.AddItem(opt.recentBucket, sug)
		}
	}
	return text
}

func PressEnterToContinue() {
	fmt.Print("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func PressEnterToContinueWithHint(hint string) {
	fmt.Print(hint)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func fileDesc(file os.FileInfo) string {
	tp := PromptTypeFile
	if file.IsDir() {
		tp = PromptTypeDir
	}
	return fmt.Sprintf("%s mod:%s", tp, file.ModTime().Format("2006-01-02 15:04:05"))
}
