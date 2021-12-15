package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestRun(t *testing.T) {

	args := []string{
		"3.4.1",
		"2021-12-24",
	}

	targetFile1 := createTempFile(t, `version=v1.0.0, date=2021-11-23`)
	defer os.Remove(targetFile1)
	targetFile2 := createTempFile(t, `version=v1.0.0, version=v2.2.1`)
	defer os.Remove(targetFile2)
	targetFile3 := createTempFile(t, `
<version>
  <major>1</major>
  <minor>0</minor>
  <revision>2</revision>
</version>`)
	defer os.Remove(targetFile3)

	config := fmt.Sprintf(`
	{
		"values" : [
			{ 
				"name" : "version",
				"regex" : "^(?P<major>[0-9]+)\\.(?P<minor>[0-9]+)\\.(?P<revision>[0-9]+)$"
			},
			{
				"name" : "date"
			}
		],
		"targets" : [
			{
				"files" : [
					"%s",
					"%s"
				],
				"embeddeds" : [
					{
						"regex" : "version=v[0-9]+\\.[0-9]+\\.[0-9]+",
						"replacement" : "version=v{{.version}}"
					},
					{
						"regex" : "date=[0-9\\-]+",
						"replacement" : "date={{.date}}"
					}
				]
			},
			{
				"files" : [
					"%s"
				],
				"embeddeds" : [
					{
						"regex" : "<major>[0-9]+</major>",
						"replacement" : "<major>{{.major}}</major>"
					},
					{
						"regex" : "<minor>[0-9]+</minor>",
						"replacement" : "<minor>{{.minor}}</minor>"
					},
					{
						"regex" : "<revision>[0-9]+</revision>",
						"replacement" : "<revision>{{.revision}}</revision>"
					}
				]
			}
		]
	}`,
		strings.ReplaceAll(targetFile1, `\`, `\\`),
		strings.ReplaceAll(targetFile2, `\`, `\\`),
		strings.ReplaceAll(targetFile3, `\`, `\\`))

	configFile := createTempFile(t, config)
	defer os.Remove(configFile)

	w := &bytes.Buffer{}
	err := run(configFile, args, w)
	if err != nil {
		t.Fatalf("failed test\n%+v", err)
	}

	{
		before := readString(t, targetFile1)
		if before != `version=v3.4.1, date=2021-12-24` {
			t.Fatal("failed test\n", before)
		}
	}
	{
		before := readString(t, targetFile2)
		if before != `version=v3.4.1, version=v3.4.1` {
			t.Fatal("failed test\n", before)
		}
	}
	{
		before := readString(t, targetFile3)
		if before != `
<version>
  <major>3</major>
  <minor>4</minor>
  <revision>1</revision>
</version>` {
			t.Fatal("failed test\n", before)
		}
	}

	output := w.String()
	// Embedded values
	if !strings.Contains(output, "version=v3.4.1") {
		t.Fatal("failed test\n", output)
	}
	if !strings.Contains(output, "date=2021-12-24") {
		t.Fatal("failed test\n", output)
	}
	if !strings.Contains(output, "<major>3</major>") {
		t.Fatal("failed test\n", output)
	}
	if !strings.Contains(output, "<minor>4</minor>") {
		t.Fatal("failed test\n", output)
	}
	if !strings.Contains(output, "<revision>1</revision>") {
		t.Fatal("failed test\n", output)
	}

	// Files
	if !strings.Contains(output, fmt.Sprintf(`[U] %s`, targetFile1)) {
		t.Fatal("failed test\n", output)
	}
	if !strings.Contains(output, fmt.Sprintf(`[U] %s`, targetFile2)) {
		t.Fatal("failed test\n", output)
	}
	if !strings.Contains(output, fmt.Sprintf(`[U] %s`, targetFile3)) {
		t.Fatal("failed test\n", output)
	}
}

func TestRun_unchanged(t *testing.T) {

	args := []string{
		"3.4.1",
	}

	targetFile1 := createTempFile(t, `version=v1.0.0`)
	defer os.Remove(targetFile1)
	targetFile2 := createTempFile(t, `date=2021-11-24`)
	defer os.Remove(targetFile2)

	config := fmt.Sprintf(`
	{
		"values" : [
			{ 
				"name" : "version",
				"regex" : "^(?P<major>[0-9]+)\\.(?P<minor>[0-9]+)\\.(?P<revision>[0-9]+)$"
			}
		],
		"targets" : [
			{
				"files" : [
					"%s",
					"%s"
				],
				"embeddeds" : [
					{
						"regex" : "version=v[0-9]+\\.[0-9]+\\.[0-9]+",
						"replacement" : "version=v{{.version}}"
					}
				]
			}
		]
	}`,
		strings.ReplaceAll(targetFile1, `\`, `\\`),
		strings.ReplaceAll(targetFile2, `\`, `\\`))

	configFile := createTempFile(t, config)
	defer os.Remove(configFile)

	w := &bytes.Buffer{}
	err := run(configFile, args, w)
	if err != nil {
		t.Fatalf("failed test\n%+v", err)
	}

	{
		before := readString(t, targetFile1)
		if before != `version=v3.4.1` {
			t.Fatal("failed test\n", before)
		}
	}
	{
		before := readString(t, targetFile2)
		if before != `date=2021-11-24` {
			t.Fatal("failed test\n", before)
		}
	}

	output := w.String()
	// Embedded values
	if !strings.Contains(output, "version=v3.4.1") {
		t.Fatal("failed test\n", output)
	}

	// Files
	if !strings.Contains(output, fmt.Sprintf(`[U] %s`, targetFile1)) {
		t.Fatal("failed test\n", output)
	}
	if !strings.Contains(output, fmt.Sprintf(`[-] %s`, targetFile2)) {
		t.Fatal("failed test\n", output)
	}
}

func TestRun_loadConfigError(t *testing.T) {

	args := []string{}

	config := ""

	configFile := createTempFile(t, config)
	defer os.Remove(configFile)

	w := &bytes.Buffer{}
	err := run(configFile, args, w)
	if err.Error() != "failed to load the config file: unexpected end of JSON input" {
		t.Fatalf("failed test\n%+v", err)
	}
}

func TestRun_valuesError(t *testing.T) {

	args := []string{
		"3.4.1",
	}

	targetFile1 := createTempFile(t, `version=v1.0.0`)
	defer os.Remove(targetFile1)

	config := fmt.Sprintf(`
	{
		"values" : [
			{ 
				"name" : "val1"
			},
			{ 
				"name" : "val2"
			}
		],
		"targets" : [
			{
				"files" : [
					"%s"
				],
				"embeddeds" : [
					{
						"regex" : "version=v[0-9]+",
						"replacement" : "version=v{{.val1}}"
					}
				]
			}
		]
	}`,
		strings.ReplaceAll(targetFile1, `\`, `\\`))

	configFile := createTempFile(t, config)
	defer os.Remove(configFile)

	w := &bytes.Buffer{}
	err := run(configFile, args, w)
	if err.Error() != "argument must be 2 arguments" {
		t.Fatalf("failed test\n%+v", err)
	}
}

func TestRun_buildReplaceRulesError(t *testing.T) {

	args := []string{
		"3.4.1",
	}

	targetFile1 := createTempFile(t, `version=v1.0.0`)
	defer os.Remove(targetFile1)

	config := fmt.Sprintf(`
	{
		"values" : [
			{ 
				"name" : "val1"
			}
		],
		"targets" : [
			{
				"files" : [
					"%s"
				],
				"embeddeds" : [
					{
						"regex" : "version=v[0-9",
						"replacement" : "version=v{{.val1}}"
					}
				]
			}
		]
	}`,
		strings.ReplaceAll(targetFile1, `\`, `\\`))

	configFile := createTempFile(t, config)
	defer os.Remove(configFile)

	w := &bytes.Buffer{}
	err := run(configFile, args, w)
	if err.Error() != "'version=v[0-9' in embeddeds.regex is an invalid value: error parsing regexp: missing closing ]: `[0-9`" {
		t.Fatalf("failed test\n%+v", err)
	}
}

func TestRun_replaceError(t *testing.T) {

	args := []string{
		"3.4.1",
	}

	targetFile1 := createTempFile(t, `version=v1.0.0`)
	defer os.Remove(targetFile1)

	config := fmt.Sprintf(`
	{
		"values" : [
			{ 
				"name" : "val1"
			}
		],
		"targets" : [
			{
				"files" : [
					"%s"
				],
				"embeddeds" : [
					{
						"regex" : "version=v[0-9]+",
						"replacement" : "version=v{{.val1}}"
					}
				]
			}
		]
	}`,
		// file not found
		strings.ReplaceAll(targetFile1+"xxxx", `\`, `\\`))

	configFile := createTempFile(t, config)
	defer os.Remove(configFile)

	w := &bytes.Buffer{}
	err := run(configFile, args, w)
	pathErr := errors.Cause(err).(*os.PathError)
	if pathErr.Path != targetFile1+"xxxx" || pathErr.Op != "open" {
		t.Fatal("failed test\n", err)
	}
}

func TestRun_executeTemplateError(t *testing.T) {

	args := []string{
		"3.4.1",
	}

	targetFile1 := createTempFile(t, `version=v1.0.0`)
	defer os.Remove(targetFile1)

	config := fmt.Sprintf(`
	{
		"values" : [
			{ 
				"name" : "val1"
			}
		],
		"targets" : [
			{
				"files" : [
					"%s"
				],
				"embeddeds" : [
					{
						"regex" : "version=v[0-9]+",
						"replacement" : "version=v{{.val1}"
					}
				]
			}
		]
	}`,
		strings.ReplaceAll(targetFile1, `\`, `\\`))

	configFile := createTempFile(t, config)
	defer os.Remove(configFile)

	w := &bytes.Buffer{}
	err := run(configFile, args, w)
	if err.Error() != "'version=v{{.val1}' in embeddeds.replacement is an invalid value: template: template:1: unexpected \"}\" in operand" {
		t.Fatalf("failed test\n%+v", err)
	}
}

func TestReplace(t *testing.T) {

	contents := "version: 1, date: 2021-12-14"

	file := createTempFile(t, contents)
	defer os.Remove(file)

	replaceRules := []ReplaceRule{
		{
			Regex:       regexp.MustCompile(`version: ([0-9]+)`),
			Replacement: "version: 2",
		},
		{
			Regex:       regexp.MustCompile(`date: ([0-9\-]+)`),
			Replacement: "date: 2021-12-24",
		},
	}

	result, err := replace(file, replaceRules)
	if err != nil {
		t.Fatalf("failed test\n%+v", err)
	}

	if !result {
		t.Fatal("failed test\n", result)
	}

	before := readString(t, file)
	if before != "version: 2, date: 2021-12-24" {
		t.Fatal("failed test\n", before)
	}
}

func TestBuildReplaceRules(t *testing.T) {

	embeddeds := []Embedded{
		{
			RegexStr:    "val1=(.+)",
			Replacement: "val1={{.val1}}",
		},
		{
			RegexStr:    "val2=(.+)",
			Replacement: "val2={{.val2}}",
		},
	}

	values := map[string]string{
		"val1": "a",
		"val2": "b",
	}

	result, err := buildReplaceRules(embeddeds, values)
	if err != nil {
		t.Fatalf("failed test\n%+v", err)
	}

	expect := []ReplaceRule{
		{
			Regex:       regexp.MustCompile("val1=(.+)"),
			Replacement: "val1=a",
		},
		{
			Regex:       regexp.MustCompile("val2=(.+)"),
			Replacement: "val2=b",
		},
	}

	if !reflect.DeepEqual(result, expect) {
		t.Fatal("failed test\n", result)
	}
}

func TestExecuteTemplate(t *testing.T) {

	values := map[string]string{
		"val1": "a",
		"val2": "b",
	}

	templStr := "val1={{.val1}}, val2={{.val2}}"

	result, err := executeTemplate(templStr, values)
	if err != nil {
		t.Fatalf("failed test\n%+v", err)
	}

	if result != "val1=a, val2=b" {
		t.Fatal("failed test\n", result)
	}
}

func TestValues(t *testing.T) {

	args := []string{
		"10.0.3",
		"x",
	}
	valueConfigs := []Value{
		{
			Name:     "version",
			RegexStr: "^(?P<major>[0-9]+)\\.(?P<minor>[0-9]+)\\.(?P<revision>[0-9]+)$",
		},
		{
			Name: "val2",
		},
	}

	result, err := values(args, valueConfigs)
	if err != nil {
		t.Fatalf("failed test\n%+v", err)
	}

	expect := map[string]string{
		"version":  "10.0.3",
		"major":    "10",
		"minor":    "0",
		"revision": "3",
		"val2":     "x",
	}

	if !reflect.DeepEqual(result, expect) {
		t.Fatal("failed test\n", result)
	}
}

func TestValues_invalidRegex(t *testing.T) {

	args := []string{
		"10.0.3",
	}
	valueConfigs := []Value{
		{
			Name:     "version",
			RegexStr: "^(",
		},
	}

	_, err := values(args, valueConfigs)
	if err.Error() != "'^(' in values.regex is an invalid value: error parsing regexp: missing closing ): `^(`" {
		t.Fatalf("failed test\n%+v", err)
	}
}

func TestValues_unmatchRegex(t *testing.T) {

	args := []string{
		"10.0.3",
	}
	valueConfigs := []Value{
		{
			Name:     "version",
			RegexStr: "^[0-9]+$",
		},
	}

	_, err := values(args, valueConfigs)
	if err.Error() != "'10.0.3' does not match the regular expression: ^[0-9]+$" {
		t.Fatalf("failed test\n%+v", err)
	}
}

func TestLoadConfig(t *testing.T) {

	config := `
{
    "values" : [
        { 
            "name" : "version",
            "regex" : "^(?P<major>[0-9]+)\\.(?P<minor>[0-9]+)\\.(?P<revision>[0-9]+)$"
        },
        {
            "name" : "value2"
        }
    ],
    "targets" : [
        {
            "files" : [
                "version.properties",
                "version2.properties"
            ],
            "embeddeds" : [
                {
                    "regex" : "version=v[0-9]+\\.[0-9]+\\.[0-9]+",
                    "replacement" : "version=v{{.version}}"
                }
            ]
        },
        {
            "files" : [
                "version.xml"
            ],
            "embeddeds" : [
                {
                    "regex" : "<major>[0-9]+</major>",
                    "replacement" : "<major>{{.major}}</major>"
                },
                {
                    "regex" : "<minor>[0-9]+</minor>",
                    "replacement" : "<minor>{{.major}}</minor>"
                },
                {
                    "regex" : "<revision>[0-9]+</revision>",
                    "replacement" : "<revision>{{.revision}}</revision>"
                }
            ]
        }
    ]
}
`

	file := createTempFile(t, config)
	defer os.Remove(file)

	result, err := loadConfig(file)
	if err != nil {
		t.Fatalf("failed test\n%+v", err)
	}

	expect := &Config{
		Values: []Value{
			{
				Name:     "version",
				RegexStr: "^(?P<major>[0-9]+)\\.(?P<minor>[0-9]+)\\.(?P<revision>[0-9]+)$",
			},
			{
				Name: "value2",
			},
		},
		Targets: []Target{
			{
				Files: []string{
					"version.properties",
					"version2.properties",
				},
				Embeddeds: []Embedded{
					{
						RegexStr:    "version=v[0-9]+\\.[0-9]+\\.[0-9]+",
						Replacement: "version=v{{.version}}",
					},
				},
			},
			{
				Files: []string{
					"version.xml",
				},
				Embeddeds: []Embedded{
					{
						RegexStr:    "<major>[0-9]+</major>",
						Replacement: "<major>{{.major}}</major>",
					},
					{
						RegexStr:    "<minor>[0-9]+</minor>",
						Replacement: "<minor>{{.major}}</minor>",
					},
					{
						RegexStr:    "<revision>[0-9]+</revision>",
						Replacement: "<revision>{{.revision}}</revision>",
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(result, expect) {
		t.Fatal("failed test\n", result)
	}
}

func TestLoadConfig_invalidFormat(t *testing.T) {

	config := `
{
    "value" : [
        { 
            "name" : "version",
            "regex" : "^(?P<major>[0-9]+)\\.(?P<minor>[0-9]+)\\.(?P<revision>[0-9]+)$"
        }
    ]
}
`

	file := createTempFile(t, config)
	defer os.Remove(file)

	_, err := loadConfig(file)
	if err.Error() != "invalid format" {
		t.Fatalf("failed test\n%+v", err)
	}
}

func TestLoadConfig_fileNotfound(t *testing.T) {

	file := createTempFile(t, "")
	defer os.Remove(file)

	_, err := loadConfig(file + "xxxx")
	pathErr := errors.Cause(err).(*os.PathError)
	if pathErr.Path != file+"xxxx" || pathErr.Op != "open" {
		t.Fatal("failed test\n", err)
	}
}

func createTempFile(t *testing.T, content string) string {

	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal("craete file failed\n", err)
	}
	defer tempFile.Close()

	_, err = tempFile.Write([]byte(content))
	if err != nil {
		t.Fatal("write file failed\n", err)
	}

	return tempFile.Name()
}

func readString(t *testing.T, file string) string {

	bo, err := os.ReadFile(file)
	if err != nil {
		t.Fatal("read failed\n", err)
	}

	return string(bo)
}
