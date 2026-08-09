export type Expose = "all" | "none" | "export";
export type JsDoc = "none" | "basic" | "extended";
export type FunctionOptions = "fail" | "comment" | "hide";

/**
 * Generator options. The keys mirror the `Config` type of
 * ts-json-schema-generator 2.x; each one maps to a flag of the native CLI.
 * Passing any other key throws.
 */
export interface Config {
    /** Glob of source files to read, supporting `*` and `**`. */
    path?: string;
    /** Type name(s) to generate a schema for; `"*"` for all exposed types. */
    type?: string | string[];
    /** Path to the tsconfig.json to compile with. Defaults to the nearest one above the working directory. */
    tsconfig?: string;
    /** `$id` of the generated schema. */
    schemaId?: string;
    /** Which types get their own definition. Default `"export"`. */
    expose?: Expose;
    /** How much of the JSDoc annotations to read. Default `"extended"`. */
    jsDoc?: JsDoc;
    /** What to do with function types. Default `"comment"`. */
    functions?: FunctionOptions;
    /** Emit `markdownDescription` alongside `description`. Implies `jsDoc: "extended"`. */
    markdownDescription?: boolean;
    /** Emit the raw JSDoc comment as `fullDescription`. Implies `jsDoc: "extended"`. */
    fullDescription?: boolean;
    /** Sort object properties. Default `true`. */
    sortProps?: boolean;
    /** Emit a top-level `$ref` definition. Default `true`. */
    topRef?: boolean;
    /** Disallow additional items on tuples. Default `false`. */
    strictTuples?: boolean;
    /** Skip type checking, which is faster but reports no diagnostics. Default `false`. */
    skipTypeCheck?: boolean;
    /** Percent-encode characters in `$ref` values. Default `true`. */
    encodeRefs?: boolean;
    /** Allow additional properties on objects without an index signature. Default `false`. */
    additionalProperties?: boolean;
    /** Extra JSDoc tags to copy into the schema as validation keywords. */
    extraTags?: string[];
    /** Accepted and ignored: the schema is returned as an object. */
    minify?: boolean;
    /** Only `"json-schema"` is supported. */
    discriminatorType?: "json-schema";
}

/**
 * Generates a JSON schema by running the native CLI, resolving with the parsed
 * schema. Rejects with an `Error` carrying the CLI's `stderr` when it fails.
 */
export function generateSchema(config: Config): Promise<Record<string, unknown>>;

/** {@link generateSchema}, synchronously. Throws instead of rejecting. */
export function generateSchemaSync(config: Config): Record<string, unknown>;
