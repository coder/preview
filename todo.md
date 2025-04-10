# TODO:

## Errors

- Dynamic parameter names? A parameter's name can change. Should we throw a warning if it is not static?
- [18](https://github.com/coder/preview/issues/18) `terraform init` not run before a `preview` fails to load a module. Should this prevent a preview?
- [29](https://github.com/coder/preview/issues/29) Providing input values for parameters that does not exist should return a warning.
- Unresolved modules should throw an error/warning that the preview is incomplete.
- 2 options with the same value should throw an error.

## Security

- [19](https://github.com/coder/preview/issues/19) Ensure local disk is not accessible from terraform
- [20](https://github.com/coder/preview/issues/20) Ensure no remote http requests for module fetching

## Performance

- Plan hook replaces the same context for every block in a module. This work is duplicated and could be trimmed down.
- [21](https://github.com/coder/preview/issues/21) Ensure no panics can occur during a preview.
- websocket should use shared cache. 2 template websockets using the same files should not load the files into memory twice. 
- Make a template with 10,000 options. Test the performance.
- Add a parameter with 50 options to the demo template.
  - searchable as well
- Demo template should be multi-select for the IDE selector.

## Features

- Allow a "force submit" to bypass any `preview` errors. This would defer to the terraform errors (basically the status quo today)
- [22](https://github.com/coder/preview/issues/22) Errors during the parsing should be reported.
- Errors during the hooks should be reported.
- Interactive shell to debug references

## Documentation

- Diagram the terraform flow
  - When `data` blocks are fetched
  - `resource` blocks are unavailable
- Enumerate common error cases


## Verification

- Nested blocks (within 1 module) should have correct context set via plan files. Since plan files are set on the parent, the parent of a sub-block is the incorrect level for a context.
 - This might be already correct

## Debt

- [23](https://github.com/coder/preview/issues/23) Implement `validation` blocks with a common code component to be reused by terraform provider?
- Parameter values/defaults are only `string` types. 
- Parameter groups/sections. Required?
- Add a custom linter to prevent `cty.Type == cty.Type`. Use `cty.Type.Equals(cty.Type)` instead.

## Upstream work

- [24](https://github.com/coder/preview/issues/24)How will the hooks work if they cannot be merged upstream? Alternative?
  - Load in plan state
  - Semantics for parameter coder blocks

## Backward compatibility

- Omitting `type` behavior, is there a default?
- Backwards compatibility form controls. Default is radio vs dropdown. 

## Bugs

- [25](https://github.com/coder/preview/issues/25) Submodule references ignored in `count` meta arguments (and dynamic blocks)?
  - https://github.com/aquasecurity/trivy/pull/8479 

## Websocket demo

- Cleanup errors and directory handling code. DRY