package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
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
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

type Target struct {
	Files     []string   `json:"files"`
	Embeddeds []Embedded `json:"embeddeds"`
}

type Embedded struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
}

type ReplaceRule struct {
	Regex       *regexp.Regexp
	Replacement string
}

func main() {

	var configPath string
	var targetDirPath string
	var help bool

	flag.StringVarP(&configPath, "config", "c", "emv.json", "Config file path.")
	flag.StringVarP(&targetDirPath, "target", "t", "", "The base directory to search for target files. If not specified, it is the same directory as the config file.")
	flag.BoolVarP(&help, "help", "h", false, "Help.")
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
		usage(os.Stderr)
		os.Exit(1)
	}

	if targetDirPath == "" {
		targetDirPath = filepath.Dir(configPath)
	}

	err := run(configPath, flag.Args(), targetDirPath, os.Stdout)
	if err != nil {
		fmt.Println("\nError: ", err)
		os.Exit(1)
	}
}

func usage(w io.Writer) {

	fmt.Fprintf(w, "emv v%s (%s)\n\n", Version, Commit)
	fmt.Fprintf(w, "Usage: emv [-c CONFIG] [-t TARGET] VALUE1 ... \n\nFlags\n")
	flag.CommandLine.SetOutput(w)
	flag.PrintDefaults()
}

func run(configPath string, args []string, targetDirPath string, w io.Writer) error {

	config, err := loadConfig(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to load the config file")
	}

	values, err := values(args, config.Values)
	if err != nil {
		return err
	}

	for _, target := range config.Targets {

		replaceRules, err := buildReplaceRules(target.Embeddeds, values)
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "Embedded values:\n")
		for _, replaceRule := range replaceRules {
			fmt.Fprintf(w, "  %s\n", replaceRule.Replacement)
		}

		fmt.Fprintf(w, "Files: ([U] Updated, [-] None)\n")
		for _, file := range target.Files {

			targetFile := file
			if !filepath.IsAbs(targetFile) && targetDirPath != "" {
				targetFile = filepath.Join(targetDirPath, file)
			}

			replaced, err := replace(targetFile, replaceRules)
			if err != nil {
				return err
			}

			var changeFlag string
			if replaced {
				changeFlag = "[U]"
			} else {
				changeFlag = "[-]"
			}

			fmt.Fprintf(w, "  %s %s\n", changeFlag, file)
		}
		fmt.Println()
	}

	return nil
}

func replace(file string, replaceRules []ReplaceRule) (bool, error) {

	content, err := os.ReadFile(file)
	if err != nil {
		return false, errors.WithStack(err)
	}

	before := string(content)
	replaced := string(content)

	for _, replaceRule := range replaceRules {
		replaced = replaceRule.Regex.ReplaceAllString(replaced, replaceRule.Replacement)
	}

	if before == replaced {
		return false, nil
	}

	return true, os.WriteFile(file, []byte(replaced), 0666)
}

func buildReplaceRules(embeddeds []Embedded, values map[string]string) ([]ReplaceRule, error) {

	replaceRules := []ReplaceRule{}

	for _, emembedded := range embeddeds {

		regexp, err := regexp.Compile(emembedded.Pattern)
		if err != nil {
			return nil, errors.Wrapf(err, "'%s' in embeddeds-pattern is an invalid value", emembedded.Pattern)
		}

		replacement, err := executeTemplate(emembedded.Replacement, values)
		if err != nil {
			return nil, errors.Wrapf(err, "'%s' in embeddeds-replacement is an invalid value", emembedded.Replacement)
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
		return "", errors.WithStack(err)
	}

	w := &strings.Builder{}
	if err := templ.Execute(w, values); err != nil {
		return "", errors.WithStack(err)
	}

	return w.String(), nil
}

func values(args []string, valueConfigs []Value) (map[string]string, error) {

	values := map[string]string{}

	if len(args) != len(valueConfigs) {
		return nil, errors.Errorf("argument must be %d arguments", len(valueConfigs))
	}

	for i, valueConfig := range valueConfigs {
		values[valueConfig.Name] = args[i]

		if valueConfig.Pattern != "" {

			regexp, err := regexp.Compile(valueConfig.Pattern)
			if err != nil {
				return nil, errors.Wrapf(err, "'%s' in values-pattern is an invalid value", valueConfig.Pattern)
			}

			match := regexp.FindStringSubmatch(args[i])
			if match == nil {
				return nil, errors.Errorf("'%s' does not match the pattern: %s", args[i], valueConfig.Pattern)
			}

			for i, name := range regexp.SubexpNames() {
				if i != 0 && name != "" {
					values[name] = match[i]
				}
			}
		}
	}

	return values, nil
}

func loadConfig(path string) (*Config, error) {

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(config.Targets) == 0 || len(config.Values) == 0 {
		return nil, errors.Errorf("invalid format")
	}

	return &config, nil
}
