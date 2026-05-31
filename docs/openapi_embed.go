package docs

import _ "embed"

// OpenAPIYAML is the runtime copy of docs/openapi.yaml.
//
//go:embed openapi.yaml
var OpenAPIYAML []byte
