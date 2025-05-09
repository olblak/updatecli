package precommit

var (
	// manifestTemplate is the Go template used to generate Fleet manifests
	manifestTemplate string = `name: '{{ .ManifestName }}'
{{- if .ActionID }}
actions:
  '{{ .ActionID }}':
    title: '{{ .TargetName }}'
{{ end }}

sources:
  '{{ .SourceID }}':
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
      url: '{{ .SourceScmUrl }}'
{{- if .Digest }}
  '{{ .SourceID }}_digest':
    name: '{{ .SourceName }}'
    kind: '{{ .SourceKind }}'
    spec:
      versionfilter:
        kind: '{{ .SourceVersionFilterKind }}'
        pattern: '{{ .SourceVersionFilterPattern }}'
      url: '{{ .SourceScmUrl }}'
      key: 'hash'
{{- end }}

targets:
  '{{ .TargetID }}':
    name: '{{ .TargetName }}'
    kind: yaml
{{- if .ScmID }}
    scmid: '{{ .ScmID }}'
{{ end }}
    sourceid: '{{ .SourceID }}{{if .Digest}}_digest{{end}}'
    spec:
      file: '{{ .File }}'
      key: "{{ .TargetKey }}"
      engine: '{{ .TargetEngine }}'
`
)
