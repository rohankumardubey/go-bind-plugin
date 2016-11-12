package cli

var generateFileTemplate = `package {{.OutputPackage}}
// Autogenerated by github.com/wendigo/go-bind-plugin on {{.Build.Date}}, do not edit!
// Command: {{.Build.Command}}
//
// Plugin {{.Config.PluginPath}} info:
// - package: {{.Plugin.Package}}
// - size: {{.Plugin.Size}} bytes
// - sha256: {{.Plugin.Sha256}}

{{ $imports := .Plugin.Imports }}
import ({{ range $imports }}
  {{.}}{{end}}
)
{{$useVarReference := .Config.DereferenceVariables|not}}{{$pluginPackage := .Plugin.Package}}{{$receiver := .Config.OutputName}}{{$interface := .Config.AsInterface}}
// {{.Config.OutputName}} wraps symbols (functions and variables) exported by plugin {{.Plugin.Package}}
//
// See docs at https://godoc.org/{{$pluginPackage}}
type {{if $interface}}_{{end}}{{.Config.OutputName}} struct {
  // Exported functions
  {{range .Plugin.Functions}}_{{ .Name }} {{ .Signature }}
  {{end}}
  {{if .Config.HideVariables | not }}
  // Exported variables (public references)
  {{range .Plugin.Variables}}
  // See docs at https://godoc.org/{{$pluginPackage}}#{{ .Name }}
  {{ .Name }} {{if $useVarReference}}*{{end}}{{ .Signature }}{{end}}
  {{end}}
}

{{if $interface}}
// {{.Config.OutputName}} wraps functions exported by plugin {{.Plugin.Package}}
//
// See docs at https://godoc.org/{{$pluginPackage}}
type {{.Config.OutputName}} interface {
  // Exported functions
  {{range .Plugin.Functions}}{{ .Name }} {{ .TrimmedSignature }}
  {{end}}
}
{{end}}

{{range .Plugin.Functions}}
// {{.Name}} function was exported from plugin {{$pluginPackage}} symbol '{{.Name}}'
//
// See docs at https://godoc.org/{{$pluginPackage}}#{{.Name}}
func (p *{{if $interface}}_{{end}}{{$receiver}}) {{.Name}}{{.TrimmedSignature}} {
  {{ if .ReturnsVoid | not }}return {{ end }}p._{{ .Name }}({{ .ArgumentsCall }})
}
{{end}}

// String returnes textual representation of the wrapper. It provides info on exported symbols and variables.
func (p *{{if $interface}}_{{end}}{{$receiver}}) String() string {
  var lines []string
  lines = append(lines, "{{if $interface}}Interface{{else}}Struct{{end}} {{.Config.OutputName}}:")
  lines = append(lines, "\t- Generated on: {{.Build.Date}}")
  lines = append(lines, "\t- Command: {{.Build.Command}}")
  lines = append(lines, "\nPlugin info:")
  lines = append(lines, "\t- package: {{$pluginPackage}}")
  lines = append(lines, "\t- sha256 sum: {{.Plugin.Sha256}}")
  lines = append(lines, "\t- size: {{.Plugin.Size}} bytes")
  lines = append(lines, "\nExported functions ({{.Plugin.Functions|len}}):")
  {{ range .Plugin.Functions }}lines = append(lines, "\t- {{.Name}} {{ .Signature }}")
  {{ end }}
  {{ if .Config.HideVariables | not }}
  {{ if .Plugin.Variables }}
  lines = append(lines, "\nExported variables ({{.Plugin.Variables|len}}):")
  {{ range .Plugin.Variables }}lines = append(lines, "\t- {{.Name}} {{ .Signature }}")
  {{ end }}
  {{end}}
  {{end}}
  {{$imports := .Plugin.NamedImports}}
  {{if $imports}}
  lines = append(lines, "\nPlugin imports:"){{ range $key, $val := $imports }}
  lines = append(lines, "\t- {{ $val }} as {{ $key }}")
  {{end}}
  {{end}}
  return strings.Join(lines, "\n")
}

{{if $interface}}
var (
  _ {{.Config.OutputName }} = (*_{{.Config.OutputName}})(nil)
)
{{end}}

// Bind{{.Config.OutputName}} loads plugin from the given path and binds {{if .Config.AsInterface}}functions{{else}}symbols (variables and functions){{end}}
// to the {{if .Config.AsInterface}}struct implementing {{.Config.OutputName}} interface{{else}}{{.Config.OutputName}} struct{{end}}. {{if .Config.HideVariables | not}}{{if .Config.DereferenceVariables}}All variables are derefenenced. {{end}}{{end}}
{{ if .Config.CheckSha256 }}// When plugin is loaded sha256 checksum is computed and checked against precomputed once. On mismatch error is returned.
{{end}}func Bind{{.Config.OutputName}}(path string) ({{if.Config.AsInterface|not}}*{{end}}{{.Config.OutputName}}, error) {
    p, err := plugin.Open(path)

    if err != nil {
      return nil, fmt.Errorf("could not open plugin: %s", err)
    }
    {{ if .Config.CheckSha256 }}
    fileChecksum := func(path string) (string, error) {
    	hasher := sha256.New()

    	file, err := os.Open(path)

    	if err != nil {
    		return "", err
    	}
    	defer file.Close()

    	if _, err := io.Copy(hasher, file); err != nil {
    		return "", err
    	}

    	return hex.EncodeToString(hasher.Sum(nil)), nil
    }

    checksum, err := fileChecksum(path)
    if err != nil {
      return nil, fmt.Errorf("could not calculate file %s checksum", path)
    }

    if checksum != "{{.Plugin.Sha256}}" {
      return nil, fmt.Errorf("SHA256 checksum mismatch (expected: {{.Plugin.Sha256}}, actual: %s)", checksum)
    }{{ end }}

    ret := new({{if .Config.AsInterface}}_{{end}}{{.Config.OutputName}})
    {{range .Plugin.Functions}}
    func{{ .Name }}, err := p.Lookup("{{ .Name }}")
    if err != nil {
      return nil, fmt.Errorf("could not import function '{{ .Name }}', symbol not found: %s", err)
    }

    if typed, ok := func{{ .Name }}.({{ .Signature }}); ok {
      ret._{{ .Name }} = typed
    } else {
      return nil, fmt.Errorf("could not import function '{{ .Name }}', incompatible types '{{ .Signature }}' and '%s'", reflect.TypeOf(func{{ .Name }}))
    }
    {{end}}
    {{if .Config.HideVariables|not}}
    {{range .Plugin.Variables}}
    var{{ .Name }}, err := p.Lookup("{{ .Name }}")
    if err != nil {
      return nil, fmt.Errorf("could not import variable '{{ .Name }}', symbol not found: %s", err)
    }

    if typed, ok := var{{ .Name }}.(*{{.Signature}}); ok {
      ret.{{ .Name }} = {{if $useVarReference|not}}*{{end}}typed
    } else {
      return nil, fmt.Errorf("could not import variable '{{ .Name }}', incompatible types '{{ .Signature }}' and '%s'", reflect.TypeOf(var{{ .Name }}))
    }
    {{end}}
    {{end}}

    return ret, nil
}
`
