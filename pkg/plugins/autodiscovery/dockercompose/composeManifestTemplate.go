package dockercompose

const (
	// composeManifestTemplate is the Go template used to generate
	// Docker compose manifests
	composeManifestTemplate string = `
name: 'Bump Docker Image Tag for {{ .ImageName }}'
sources:
  {{ .SourceID }}:
    name: '[{{ .ImageName }}] Get latest Docker Image Tag'
    kind: 'dockerimage'
    spec:
      image: '{{ .ImageName }}'
      tagFilter: '{{ .TagFilter }}'
      versionFilter:
        kind: '{{ .VersionFilterKind }}'
        pattern: '{{ .VersionFilterPattern }}' 
targets:
  {{ .TargetID }}:
    name: '{{ .TargetName }}'
    kind: 'yaml'
{{ if .ScmID }}
    scmid: {{ .ScmID }}
{{ end }}
{{ if .ActionID }}
    scmid: {{ .ActionID }}
{{ end }}
    spec:
      file: '{{ .TargetFile }}'
      key: '{{ .TargetKey }}'
    sourceid: '{{ .SourceID }}'
    transformers:
      - addprefix: '{{ .TargetPrefix }}'
`
)
