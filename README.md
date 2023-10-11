# WID Notifier

A tool that sends configurable email notifications for

- https://wid.cert-bund.de/portal/wid/kurzinformationen
- https://wid.lsi.bayern.de/portal/wid/warnmeldungen

## Supported Platforms

This Software only supports Linux.

# Config

Example:

```json
{
  "api_fetch_interval": 600,
  "enabled_api_endpoints": [
    "bay",
    "bund"
  ],
  "datafile": "data",
  "recipients": [
    {
      "address": "guenther@example.org",,
      "include": [
        {"classification": "kritisch"},
        {"title_contains": "jQuery"}
      ]
    }
  ],
  "smtp": {
    "from": "from@example.org",
    "host": "example.org",
    "port": 587,
    "user": "from@example.org",
    "password": "SiEhAbEnMiChInSgEsIcHtGeFiLmTdAsDüRfEnSiEnIcHt"
  },
  "template": {
    "subject": "",
    "body": ""
  }
}
```

## Filters

You must filter the notices to be sent per user. Multiple filters can be set per user and multiple criteria can be defined per filter.

```json
"include": [
  {
    "any": false,
    "title_contains": "",
    "classification": "",
    "min_basescore": 0,
    "status": "",
    "products_contain": "",
    "no_patch": ""
  },
  ...
]
```

The following filter criteria are supported. Criteria marked with `*` are for optional fields that are not supported by every API endpoint (e.g. https://wid.lsi.bayern.de) - notices from those endpoints will therefore not be included when using those filters.

### any

Includes all notices if set to `true`.

```json
{"any": true}
```

### title_contains

Include notices whose title contains this text.

```json
{"title_contains": "Denial Of Service"}
```
If set to `""`, this criteria will be ignored.

### classification

Include notices whose classification is in this list.  
Classification can be `"kritisch"`, `"hoch"`, `"mittel"` or `"niedrig"`.

```json
{"classification": "hoch"}
```
If set to `""`, this criteria will be ignored.

### min_basescore `*`

Include notices whose basescore (`0` - `100`) is >= `min_basescore`.

```json
{"min_basescore": 40}
```
This criteria will be ignored if set to `0`.

### status `*`

Include notices with this status. This is usually either `NEU` or `UPDATE`.

```json
{"status": "NEU"}
```
If set to `""`, this criteria will be ignored.

### products_contain `*`

Include notices whose product list contains this text.

```json
{"products_contain": "Debian Linux"}
```
If set to `""`, this criteria will be ignored.

### no_patch `*`

If set to `"true"`, notices where no patch is available will be included.

```json
{"no_patch": "true"}
```

If set to `"false"`, notices where no patch is available will be included.

```json
{"no_patch": "false"}
```

If set to `""`, this criteria will be ignored.

## Templates

The syntax for the mail templates is described [here](https://pkg.go.dev/text/template).

All fields from the WidNotice struct can be used.

```go
type WidNotice struct {
  Uuid string
  Name string
  Title string
  Published time.Time
  Classification string
  // optional fields (only fully supported by cert-bund)
  Basescore int // -1 = unknown
  Status string // "" = unknown
  ProductNames []string // empty = unknown
  Cves []string // empty = unknown
  NoPatch string // "" = unknown
  // metadata
  PortalUrl string
}
```

For an example, take a look at `DEFAULT_SUBJECT_TEMPLATE` and `DEFAULT_BODY_TEMPLATE` in [template.go](./template.go).
