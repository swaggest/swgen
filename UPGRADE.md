# Upgrade guide for breaking changes

## v0.5.0

 * Non-path parameters are not required by default, please add `required:"true"` tag where you have `query:"..."` to keep original behaviour.
 * `swgen_type` is removed, please use `type` and `format` tags instead.
