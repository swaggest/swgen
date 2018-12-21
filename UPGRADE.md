# Upgrade guide for breaking changes

## v0.6.0

 * Numeric (`Min*`, `Max*`) fields of [`CommonFields`](https://godoc.org/github.com/swaggest/swgen#CommonFields) became 
 pointers to allow zero value not to be omitted. This may affect your 
 [Schema Definitions](https://godoc.org/github.com/swaggest/swgen#SchemaDefinition).

## v0.5.0

 * Non-path parameters are not required by default, please add `required:"true"` tag where you have `query:"..."` to keep original behaviour.
 * `swgen_type` is removed, please use `type` and `format` tags instead.
