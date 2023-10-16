// Copyright (c) 2023 Julian Müller (ChaoticByte)

package main

import (
	"bytes"
	"text/template"
)

const DEFAULT_SUBJECT_TEMPLATE = "{{ if .Status }}{{.Status}} {{ end }}[{{.Classification}}] {{.Title}}"
const DEFAULT_BODY_TEMPLATE = `{{.Name}}
{{.PortalUrl}}

Published: {{.Published}}
{{ if gt .Basescore -1 }}Basescore: {{.Basescore}}
{{ end -}}
{{ if eq .NoPatch "true" }}There is no patch available at the moment!
{{ end }}
Affected Products:
{{ range $product := .ProductNames }}  - {{ $product }}
{{ else }}  unknown
{{ end }}
Assigned CVEs:
{{ range $cve := .Cves }}  - {{ $cve }}
{{ else }}  unknown
{{ end }}`

type MailTemplateConfig struct {
	SubjectTemplate string `json:"subject"`
	BodyTemplate string `json:"body"`
}

type MailTemplate struct {
	SubjectTemplate template.Template
	BodyTemplate template.Template
}

func (t MailTemplate) generate(notice WidNotice) (MailContent, error) {
	c := MailContent{}
	buffer := &bytes.Buffer{}
	err := t.SubjectTemplate.Execute(buffer, notice)
	if err != nil {
		return c, err
	}
	c.Subject = buffer.String()
	buffer.Truncate(0) // we can recycle our buffer
	err = t.BodyTemplate.Execute(buffer, notice)
	if err != nil {
		return c, err
	}
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
