// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"bytes"
	"text/template"
)

const DEFAULT_SUBJECT_TEMPLATE = "[{{ .Classification }}] {{ .Title }}"
const DEFAULT_BODY_TEMPLATE = `{{ if .Status }}[{{ .Status }}] {{ end }}{{ .Name }}
-> {{ .PortalUrl }}
{{- if eq .NoPatch "true" }}

No patch available!
{{- end }}
{{ if gt .Basescore -1 }}
Basescore: {{ .Basescore }}{{- end }}
Published: {{ .Published }}
{{- if .ProductNames }}

Affected Products:{{ range $product := .ProductNames }}
  - {{ $product }}
{{- end }}{{ end }}
{{- if .Cves }}

Assigned CVEs:{{ range $cve := .Cves }}
  - {{ $cve }} -> https://www.cve.org/CVERecord?id={{ $cve }}
{{- end }}{{ end }}


Sent by WidNotifier {{ .WidNotifierVersion }}
`

type TemplateData struct {
	*WidNotice
	WidNotifierVersion string
}

type MailTemplateConfig struct {
	SubjectTemplate string `json:"subject"`
	BodyTemplate string `json:"body"`
}

type MailTemplate struct {
	SubjectTemplate template.Template
	BodyTemplate template.Template
}

func (t MailTemplate) generate(data TemplateData) (MailContent, error) {
	c := MailContent{}
	buffer := &bytes.Buffer{}
	err := t.SubjectTemplate.Execute(buffer, data)
	if err != nil { return c, err }
	c.Subject = buffer.String()
	buffer.Truncate(0) // we can recycle our buffer
	err = t.BodyTemplate.Execute(buffer, data)
	if err != nil { return c, err }
	c.Body = buffer.String()
	return c, nil
}

func NewTemplateFromTemplateConfig(tc MailTemplateConfig) MailTemplate {
	subjectTemplate, err := template.New("subject").Parse(tc.SubjectTemplate)
	if err != nil {
		logger.error("Could not parse template")
		panic(err)
	}
	bodyTemplate, err := template.New("body").Parse(tc.BodyTemplate)
	if err != nil {
		logger.error("Could not parse template")
		panic(err)
	}
	return MailTemplate{
		SubjectTemplate: *subjectTemplate,
		BodyTemplate: *bodyTemplate,
	}
}
