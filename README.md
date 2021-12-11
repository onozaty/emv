# emv

Embeds a value in the file

## Config

```json
{
    "values" : [
        { 
            "name" : "version",
            "regex" : "^(P?<major>[0-9]+)\\.(P?<minor>[0-9]+)\\.(P?<revision>[0-9]+)$"
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
                },
            ]
        }
    ]
}
```