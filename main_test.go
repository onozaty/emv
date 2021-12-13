package main

import (
	"os"
	"reflect"
	"regexp"
	"testing"
)

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
		t.Fatal("failed test\n", err)
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
		t.Fatal("failed test\n", err)
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
		t.Fatal("failed test\n", err)
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
		t.Fatal("failed test\n", err)
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
		t.Fatal("failed test\n", err)
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

func createTempFile(t *testing.T, content string) string {

	tempFile, err := os.CreateTemp("", "csv")
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
