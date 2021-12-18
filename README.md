# emv

[![GitHub license](https://img.shields.io/github/license/onozaty/emv)](https://github.com/onozaty/emv/blob/main/LICENSE)
[![Test](https://github.com/onozaty/emv/actions/workflows/test.yaml/badge.svg)](https://github.com/onozaty/emv/actions/workflows/test.yaml)

`emv` is a CLI tool for embedding specified values in files.

Embeds values in the file based on the rules written in the configuration file.

## Usage

```
emv -c project/emv.json 1.0.2
```

```
Usage: emv [-c CONFIG] [-t TARGET] VALUE1 ...

Flags
  -c, --config string   Config file path. (default "emv.json")
  -t, --target string   The base directory to search for target files. If not specified, it is the same directory as the config file.
  -h, --help            Help.
```

Define the target files and embedding contents in the config file.

As an example, prepare a config file as shown below.  
The file name should be `emv.json`.

```json
{
  "values" : [
    { 
      "name" : "version"
    }
  ],
  "targets" : [
    {
      "files" : [
        "example.properties"
      ],
      "embeddeds" : [
        {
          "pattern" : "version=[0-9\\.]+",
          "replacement" : "version={{.version}}"
        }
      ]
    }
  ]
}
```

`values` defines how the input values should be handled.  
`name` gives a name to the input value.

`targets` defines the target files and embedding contents.  
`files` is the target files. By the default is to use the same directory as the configuration file as the base. The `-t` option can be used to change the base directory.

`embeddeds` defines the embedded contents.  
`pattern` will be a regular expression. Replace `pattern` with the value of `replacement`.  
`replacement` can be embedded with the value of the input value, such as `{{.name}}`.


The contents of `example.properties` are as follows.

```conf
version=1.1.2
```

Execute it in the folder where `emv.json` and `example.properties` are located.  
`2.0.0` is embedded.

```console
$ emv 2.0.0
Embedded values:
  version=2.0.0
Files: ([U] Updated, [-] None)
  [U] example.properties
```

The contents of `example.properties` have been replaced with the following.

```conf
version=2.0.0
```

## Config

```json
{
  "values" : [
    { 
      "name" : "version",
      "pattern" : "^(?P<major>[0-9]+)\\.(?P<minor>[0-9]+)\\.(?P<revision>[0-9]+)$"
    }
  ],
  "targets" : [
    {
      "files" : [
        "example.xml"
      ],
      "embeddeds" : [
        {
          "pattern" : "<major>[0-9]+</major>",
          "replacement" : "<major>{{.major}}</major>"
        },
        {
          "pattern" : "<minor>[0-9]+</minor>",
          "replacement" : "<minor>{{.minor}}</minor>"
        },
        {
          "pattern" : "<revision>[0-9]+</revision>",
          "replacement" : "<revision>{{.revision}}</revision>"
        }
      ]
    }
  ]
}
```

* `values` : The definition of the input values to be specified as arguments.
  * `name` : The name to give to the input value. This is the name to use in `replacement`.
  * `pattern` : Input value pattern. It is specified by a regular expression.<br>By writing a named group, you can name the part and it will be available in `replacement`.
* `targets` : The definition of the embedding target.
  * `files` : Target files.<br>If you specify a relative path, the default is to use the same directory as the configuration file as the base. The `-t` option can be used to change the base directory.
  * `embeddeds` : The definition of the embedded content.
    * `pattern` : The embedding position. It is specified by a regular expression.
    * `replacement` : The value to be embedded. You can use `{{.name}}` to specify the input value.


Please refer to the following for the syntax of regular expressions.

* https://pkg.go.dev/regexp/syntax

## Install

emv is implemented in golang and runs on all major platforms such as Windows, Mac OS, and Linux.  
You can download the binaries for each OS from the links below.

* https://github.com/onozaty/emv/releases/latest

## License

MIT

## Author

[onozaty](https://github.com/onozaty)
