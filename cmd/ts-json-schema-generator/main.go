// Command ts-json-schema-generator generates JSON Schema from TypeScript
// sources, mirroring ts-json-schema-generator.ts.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vega/ts-json-schema-generator-go/internal/config"
	"github.com/vega/ts-json-schema-generator-go/internal/factory"
	"github.com/vega/ts-json-schema-generator-go/internal/generator"
	"github.com/vega/ts-json-schema-generator-go/internal/schema"
)

// version is the release version; overridden at build time via
// -ldflags "-X main.version=v...".
var version = "dev"

// stringList is a repeatable string flag value. Values are not split on
// commas: type names can contain them (e.g. "Generic<A,B>").
type stringList []string

func (l *stringList) String() string { return strings.Join(*l, " ") }

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("ts-json-schema-generator", flag.ContinueOnError)

	var (
		path               string
		types              stringList
		id                 string
		tsconfig           string
		expose             string
		jsDoc              string
		markdownDesc       bool
		fullDesc           bool
		functions          string
		minify             bool
		unstable           bool
		strictTuples       bool
		noTopRef           bool
		noTypeCheck        bool
		noRefEncode        bool
		out                string
		outdir             string
		validationKeywords stringList
		additionalProps    bool
		showVersion        bool
	)

	flags.StringVar(&path, "path", "", "Source file path (glob; supports * and **)")
	flags.StringVar(&path, "p", "", "Alias for --path")
	flags.Var(&types, "type", "Type name; repeat the flag for multiple types ('*' for all)")
	flags.Var(&types, "t", "Alias for --type")
	flags.StringVar(&id, "id", "", "$id for generated schema")
	flags.StringVar(&id, "i", "", "Alias for --id")
	flags.StringVar(&tsconfig, "tsconfig", "", "Custom tsconfig.json path")
	flags.StringVar(&tsconfig, "f", "", "Alias for --tsconfig")
	flags.StringVar(&expose, "expose", "export", "Type exposing: all, none, or export")
	flags.StringVar(&expose, "e", "export", "Alias for --expose")
	flags.StringVar(&jsDoc, "jsDoc", "extended", "Read JSDoc annotations: none, basic, or extended")
	flags.StringVar(&jsDoc, "j", "extended", "Alias for --jsDoc")
	flags.BoolVar(&markdownDesc, "markdown-description", false, "Generate `markdownDescription` in addition to `description` (implies --jsDoc extended)")
	flags.BoolVar(&fullDesc, "full-description", false, "Include the full raw JSDoc comment as `fullDescription` in the schema (implies --jsDoc extended)")
	flags.StringVar(&functions, "functions", "comment", "How to handle functions: fail, comment, or hide")
	flags.BoolVar(&minify, "minify", false, "Minify generated schema")
	flags.BoolVar(&unstable, "unstable", false, "Do not sort properties")
	flags.BoolVar(&strictTuples, "strict-tuples", false, "Do not allow additional items on tuples")
	flags.BoolVar(&noTopRef, "no-top-ref", false, "Do not create a top-level $ref definition")
	flags.BoolVar(&noTypeCheck, "no-type-check", false, "Skip type checks to improve performance")
	flags.BoolVar(&noRefEncode, "no-ref-encode", false, "Do not encode references")
	flags.StringVar(&out, "out", "", "Set the output file (default: stdout)")
	flags.StringVar(&out, "o", "", "Alias for --out")
	flags.StringVar(&outdir, "outdir", "", "Write one schema file per --type to <dir>/<type>.schema.json (parses the sources once)")
	flags.Var(&validationKeywords, "validation-keywords", "Provide additional validation keywords to include; repeatable")
	flags.BoolVar(&additionalProps, "additional-properties", false, "Allow additional properties for objects with no index signature")
	flags.BoolVar(&showVersion, "version", false, "Print the version and exit")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if rest := flags.Args(); len(rest) > 0 {
		return fmt.Errorf("unexpected argument %q (flags must precede values; repeat --type for multiple types)", rest[0])
	}

	if showVersion {
		fmt.Println(version)
		return nil
	}

	if err := validateChoice("expose", expose, "all", "none", "export"); err != nil {
		return err
	}
	if err := validateChoice("jsDoc", jsDoc, "none", "basic", "extended"); err != nil {
		return err
	}
	if err := validateChoice("functions", functions, "fail", "comment", "hide"); err != nil {
		return err
	}
	// Checked via Visit rather than a non-empty value so that an explicit
	// --outdir "" is an error instead of silently falling back to stdout.
	outdirSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "outdir" {
			outdirSet = true
		}
	})
	if outdirSet {
		if err := validateOutdirFlags(outdir, out, types); err != nil {
			return err
		}
	}

	// --markdown-description and --full-description imply --jsDoc extended.
	if markdownDesc || fullDesc {
		jsDoc = "extended"
	}

	// Like the TypeScript CLI (ts.findConfigFile), fall back to the nearest
	// tsconfig.json above the working directory when --tsconfig is not given.
	if tsconfig == "" {
		tsconfig = findConfigFile()
	}

	cfg := &config.Config{
		Minify:               minify,
		Path:                 path,
		Tsconfig:             tsconfig,
		Types:                types,
		SchemaID:             id,
		Expose:               config.Expose(expose),
		TopRef:               !noTopRef,
		JSDoc:                config.JSDocMode(jsDoc),
		MarkdownDescription:  markdownDesc,
		FullDescription:      fullDesc,
		SortProps:            !unstable,
		StrictTuples:         strictTuples,
		SkipTypeCheck:        noTypeCheck,
		EncodeRefs:           !noRefEncode,
		ExtraTags:            validationKeywords,
		AdditionalProperties: additionalProps,
		DiscriminatorType:    config.DiscriminatorJSONSchema,
		Functions:            config.FunctionOptions(functions),
	}

	gen, release, err := factory.CreateGenerator(cfg)
	if err != nil {
		return err
	}
	defer release()

	// One schema file per type, all sharing the program built above: loading,
	// parsing and type-checking the project happen once for the whole set and
	// only the per-type schema walk repeats.
	if outdir != "" {
		for _, typeName := range cfg.Types {
			// The chain was built for the whole --type list, so it does not
			// know which single type each file is for.
			gen.SetTopRefName(typeName)
			output, err := generateSchema(gen, cfg, []string{typeName})
			if err != nil {
				return fmt.Errorf("generating schema for type %q: %w", typeName, err)
			}
			if err := writeSchemaFile(filepath.Join(outdir, typeName+".schema.json"), output); err != nil {
				return err
			}
		}
		return nil
	}

	output, err := generateSchema(gen, cfg, cfg.Types)
	if err != nil {
		return err
	}

	if out != "" {
		return writeSchemaFile(out, output)
	}
	_, err = os.Stdout.Write(output)
	return err
}

func generateSchema(gen *generator.SchemaGenerator, cfg *config.Config, types []string) ([]byte, error) {
	generated, err := gen.CreateSchema(types)
	if err != nil {
		return nil, err
	}
	output, err := schema.MarshalStable(generated, cfg.SortProps, cfg.Minify)
	if err != nil {
		return nil, err
	}
	// Match the reference CLI byte-for-byte: writeFileSync adds no newline,
	// while the stdout path is console.log(schemaString + "\n").
	return bytes.TrimRight(output, "\n"), nil
}

func writeSchemaFile(path string, output []byte) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, output, 0o644)
}

// validateOutdirFlags rejects the --outdir combinations whose per-type file
// names would be ambiguous, missing, or unsafe. It runs before the sources are
// parsed, so a mistake costs no compile time.
func validateOutdirFlags(outdir, out string, types []string) error {
	if outdir == "" {
		return errors.New("--outdir needs a directory path")
	}
	if out != "" {
		return errors.New("--out and --outdir are mutually exclusive")
	}
	if info, err := os.Stat(outdir); err == nil && !info.IsDir() {
		return fmt.Errorf("--outdir %q exists and is not a directory", outdir)
	}
	if len(types) == 0 {
		return errors.New("--outdir requires at least one --type: it writes one file per named type")
	}
	seen := make(map[string]bool, len(types))
	for _, typeName := range types {
		if typeName == "*" {
			return errors.New("--outdir cannot be used with --type '*'; list the types explicitly")
		}
		if seen[typeName] {
			return fmt.Errorf("duplicate --type %q: each type writes its own file under --outdir", typeName)
		}
		seen[typeName] = true
		if typeName == "" || strings.ContainsAny(typeName, `/\`+"\x00") {
			return fmt.Errorf("type %q cannot be used as a file name under --outdir: it must be non-empty and free of path separators and NUL bytes", typeName)
		}
	}
	return nil
}

func validateChoice(name, value string, choices ...string) error {
	for _, choice := range choices {
		if value == choice {
			return nil
		}
	}
	return fmt.Errorf("invalid value %q for --%s (choices: %s)", value, name, strings.Join(choices, ", "))
}

// findConfigFile walks up from the working directory looking for a
// tsconfig.json, mirroring ts.findConfigFile.
func findConfigFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, "tsconfig.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
