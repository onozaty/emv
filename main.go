package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"regexp"
	"strings"

	flag "github.com/spf13/pflag"
)

var (
	Version = "dev"
	Commit  = "none"
)

type Config struct {
	Values  []Value  `json:"values"`
	Targets []Target `json:"targets"`
}

type Value struct {
	Name     string `json:"name"`
	RegexStr string `json:"regex"`
}

type Target struct {
	Files     []string   `json:"files"`
	Embeddeds []Embedded `json:"embeddeds"`
}

type Embedded struct {
	RegexStr    string `json:"regex"`
	Replacement string `json:"replacement"`
}

type ReplaceRule struct {
	Regex       *regexp.Regexp
	Replacement string
}

func main() {

	var configPath string
	var help bool

	flag.StringVarP(&configPath, "config", "c", "emv.config", "Config file path")
	flag.BoolVarP(&help, "help", "h", false, "Help")
	flag.CommandLine.SortFlags = false
	flag.Usage = func() {
		usage(os.Stderr)
	}

	flag.Parse()

	if help {
		usage(os.Stdout)
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	values, err := values(flag.Args(), config.Values)
	if err != nil {
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}

	for _, target := range config.Targets {

		replaceRules, err := buildReplaceRules(target.Embeddeds, values)
		if err != nil {
			fmt.Println("\nError: ", err)
			os.Exit(1)
		}

		fmt.Printf("\nEmbedded values: \n")
		for _, replaceRule := range replaceRules {
			fmt.Printf("  %s\n", replaceRule.Replacement)
		}

		fmt.Printf("Files: \n")
		for _, file := range target.Files {
			if err := replace(file, replaceRules); err != nil {
				fmt.Println("\nError: ", err)
				os.Exit(1)
			}

			fmt.Printf("  %s\n", file)
		}
	}
}

func usage(w io.Writer) {

	fmt.Printf("emv v%s (%s)\n\n", Version, Commit)
	fmt.Fprint(w, "Usage: emv [-c CONFIG] VALUE1 VALUE2 ... \n\nFlags\n")
	flag.PrintDefaults()
}

func replace(file string, replaceRules []ReplaceRule) error {

	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	before := string(content)
	replaced := string(content)

	for _, replaceRule := range replaceRules {
		replaced = replaceRule.Regex.ReplaceAllString(replaced, replaceRule.Replacement)
	}

	if before != replaced {
		return fmt.Errorf("could not find the embedding position : %s", file)
	}

	return os.WriteFile(file, []byte(before), 0666)
}

func buildReplaceRules(embeddeds []Embedded, values map[string]string) ([]ReplaceRule, error) {

	replaceRules := []ReplaceRule{}

	for _, emembedded := range embeddeds {

		regexp, err := regexp.Compile(emembedded.RegexStr)
		if err != nil {
			return nil, err
		}

		replacement, err := executeTemplate(emembedded.Replacement, values)
		if err != nil {
			return nil, err
		}

		replaceRules = append(replaceRules, ReplaceRule{
			Regex:       regexp,
			Replacement: replacement,
		})
	}

	return replaceRules, nil
}

func executeTemplate(templStr string, values map[string]string) (string, error) {

	templ, err := template.New("template").Parse(templStr)
	if err != nil {
		return "", err
	}

	w := &strings.Builder{}
	if err := templ.Execute(w, values); err != nil {
		return "", err
	}

	return w.String(), nil
}

func values(args []string, valueConfigs []Value) (map[string]string, error) {

	values := map[string]string{}

	if len(args) != len(valueConfigs) {
		return nil, fmt.Errorf("argument must be %d arguments", len(valueConfigs))
	}

	for i, valueConfig := range valueConfigs {
		values[valueConfig.Name] = args[i]

		regexp, err := regexp.Compile(valueConfig.RegexStr)
		if err != nil {
			return nil, err
		}

		match := regexp.FindStringSubmatch(args[i])
		if match == nil {
			return nil, fmt.Errorf("%s does not match the regular expression: %s", args[i], valueConfig.RegexStr)
		}

		for i, name := range regexp.SubexpNames() {
			if i != 0 && name != "" {
				values[name] = match[i]
			}
		}
	}

	return values, nil
}

func loadConfig(path string) (*Config, error) {

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
