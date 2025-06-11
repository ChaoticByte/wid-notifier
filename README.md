# WidNotifier

The German [BSI](https://www.bsi.bund.de/) and [LSI Bavaria](https://lsi.bayern.de/) each have a page listing current security notices.  
This software queries the APIs of these services for new security notices and sends configurable email notifications.

## API Endpoints

|     | ID     | Portal URL                                            |
|-----|--------|-------------------------------------------------------|
| BSI | `bund` | https://wid.cert-bund.de/portal/wid/kurzinformationen |
| LSI | `bay`  | https://wid.lsi.bayern.de/portal/wid/warnmeldungen    |

## Supported Platforms

This Software only supports Linux.

# Build

To cross-compile the software for `i386`, `amd64`, `arm` and `arm64`, run `build.sh`.  
You need Go 1.22.x and git.

# Usage

```bash
./wid-notifier <configfile>
```

where `<configfile>` is the path of your configuration file. If you don't have a config file yet, the software will create an initial config at the given location. See [Configuration](#configuration) for more info.

# Configuration

Example:

```json
{
  "api_fetch_interval": 600,
  "enabled_api_endpoints": [
    "bay",
    "bund"
  ],
  "datafile": "data.json",
  "loglevel": 2,
  "lists": [
    {
      "name": "Example List",
      "recipients": ["someone@example.org"],
      "filter": [
        {"classification": "hoch", "title_contains": "Microsoft"},
        {"classification": "kritisch"}
      ]
    }
  ],
  "smtp": {
    "from": "user@localhost",
    "host": "127.0.0.1",
    "port": 587,
    "user": "user@localhost",
    "password": "change me :)"
  },
  "template": {
    "subject": "",
    "body": ""
  }
}
```

To show debug messages, set the `loglevel` to `3`.

## Filters

You define filters for notices to be sent (per recipient). Multiple filters can be set per recipient and multiple criteria can be used per filter. The configuration field for those filters is `include`. See [Configuration](#configuration) for an example.

If a notice is included is determined by the following logic:

```
{criteria, criteria, ... ALL APPLY}
OR {criteria, criteria, ... ALL APPLY}
OR ...
```

The following criteria are available. Criteria marked with * are optional fields that are not supported by every API endpoint (e.g. https://wid.lsi.bayern.de) - notices from those endpoints will therefore not be included when using those criteria in filters.

```json
"include": [
  {
    "any": false,
    "title_contains": "",
    "classification": "",
    "min_basescore": 0,
    "status": "",
    "products_contain": "",
    "no_patch": "",
    "api_endpoint": ""
  },
  ...
]
```

### any

Includes all notices if set to `true`.

```json
"any": true
```

### title_contains

Include notices whose title contains this text.

```json
"title_contains": "Denial Of Service"
```
If set to `""`, this criteria will be ignored.

### classification

Include notices whose classification is in this list.  
Classification can be `"kritisch"`, `"hoch"`, `"mittel"` or `"niedrig"`.

```json
"classification": "hoch"
```
If set to `""`, this criteria will be ignored.

### min_basescore *

Include notices whose basescore (`0` - `100`) is >= `min_basescore`.

```json
"min_basescore": 40
```
This criteria will be ignored if set to `0`.

### status *

Include notices with this status. This is usually either `NEU` or `UPDATE`.

```json
"status": "NEU"
```
If set to `""`, this criteria will be ignored.

### products_contain *

Include notices whose product list contains this text.

```json
"products_contain": "Debian Linux"
```
If set to `""`, this criteria will be ignored.

### no_patch *

If set to `"true"`, notices where no patch is available will be included.

```json
"no_patch": "true"
```

If set to `"false"`, notices where no patch is available will be included.

```json
"no_patch": "false"
```

If set to `""`, this criteria will be ignored.

### api_endpoint

Includes notices from the given [API Endpoint](#api-endpoints).

```json
"api_endpoint": "bund"
```

If set to `""`, this criteria will be ignored.

## Templates

If you don't like the default appearance of the notification mails, you can write your own templates for the mail subject and body.

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
  ApiEndpointId string
  PortalUrl string
}
```

Additionally, the field `WidNotifierVersion` holds the version of the software.

For an example, take a look at `DEFAULT_SUBJECT_TEMPLATE` and `DEFAULT_BODY_TEMPLATE` in [template.go](./template.go).
