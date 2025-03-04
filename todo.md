# TODO:

## Errors

- Dynamic parameter names? Changing names?
- `terraform init` not run before a `preview` fails to load a module. Should this prevent a preview?

## Security

- Ensure local disk is not accessible from terraform
- Ensure no remote http requests for module fetching

## Performance

- Ensure no remote http requests for module fetching
- Plan hook replaces the same context for every block in a module. This work is duplicated and could be trimmed down.
- Ensure no panics can occur during a preview.

## Features

- Allow a "force submit" to bypass any `preview` errors. This would defer to the terraform errors (basically the status quo today)
- Errors during the parsing should be reported.
- Errors during the hooks should be reported.
- Interactive shell to debug references

## Documentation

- Diagram the terraform flow
  - When `data` blocks are fetched
  - `resource` blocks are unavailable
- Enumerate common error cases


## Verification

- Nested blocks (within 1 module) should have correct context set via plan files. Since plan files are set on the parent, the parent of a sub-block is the incorrect level for a context.

## Debt

- Implement `validation` blocks with a common code component to be reused by terraform provider?
- Parameter values/defaults are only `string` types. 

## Upstream work

- How will the hooks work if they cannot be merged upstream? Alternative?
  - Load in plan state
  - Semantics for parameter coder blocks

## Backward compatibility

- Omitting `type` behavior, is there a default?

## Bugs

- Submodule references ignored in `count` meta arguments (and dynamic blocks)?
 - https://github.com/aquasecurity/trivy/pull/8479 

## Websocket demo

- Cleanup errors and directory handling code. DRY