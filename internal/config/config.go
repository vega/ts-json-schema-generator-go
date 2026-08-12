// Package config holds generator configuration, mirroring src/Config.ts.
package config

// Expose selects which types get their own definition in the output.
type Expose string

// Valid Expose values.
const (
	ExposeAll    Expose = "all"
	ExposeNone   Expose = "none"
	ExposeExport Expose = "export"
)

// JSDocMode selects how much JSDoc is carried over into the schema.
type JSDocMode string

// Valid JSDocMode values.
const (
	JSDocNone     JSDocMode = "none"
	JSDocBasic    JSDocMode = "basic"
	JSDocExtended JSDocMode = "extended"
)

// DiscriminatorType selects the encoding used for discriminated unions.
type DiscriminatorType string

// Valid DiscriminatorType values.
const (
	DiscriminatorJSONSchema DiscriminatorType = "json-schema"
	DiscriminatorOpenAPI    DiscriminatorType = "open-api"
)

// FunctionOptions selects how function types are handled.
type FunctionOptions string

// Valid FunctionOptions values.
const (
	FunctionsFail    FunctionOptions = "fail"
	FunctionsComment FunctionOptions = "comment"
	FunctionsHide    FunctionOptions = "hide"
)

// Config is the full set of generator options. Use Default for the defaults;
// the zero value is not a usable configuration.
type Config struct {
	// Path is a glob pattern for source TypeScript files to process. If not
	// provided, falls back to files from tsconfig.
	Path string
	// Types are the type names to generate schemas for; "*" means all.
	Types []string
	// Minify controls whitespace in the output JSON.
	Minify bool
	// SchemaID sets the $id property of the generated schema.
	SchemaID string
	// Tsconfig is the path to a tsconfig.json used for compilation.
	Tsconfig string
	Expose   Expose
	// TopRef wraps the root type in a $ref definition.
	TopRef bool
	JSDoc  JSDocMode
	// MarkdownDescription adds markdownDescription alongside description.
	MarkdownDescription bool
	// FullDescription includes the raw JSDoc comment as fullDescription.
	FullDescription bool
	// SortProps sorts object properties alphabetically.
	SortProps bool
	// StrictTuples disallows additional items on tuples.
	StrictTuples bool
	// SkipTypeCheck skips TypeScript semantic diagnostics.
	SkipTypeCheck bool
	// EncodeRefs URI-encodes $ref values.
	EncodeRefs bool
	// ExtraTags are additional JSDoc tag names to include in the schema.
	ExtraTags []string
	// AdditionalProperties is the default for objects without index signatures.
	AdditionalProperties bool
	DiscriminatorType    DiscriminatorType
	Functions            FunctionOptions
}

// Default returns the default configuration (DEFAULT_CONFIG in src/Config.ts).
func Default() *Config {
	return &Config{
		Expose:            ExposeExport,
		TopRef:            true,
		JSDoc:             JSDocExtended,
		SortProps:         true,
		EncodeRefs:        true,
		DiscriminatorType: DiscriminatorJSONSchema,
		Functions:         FunctionsComment,
	}
}
