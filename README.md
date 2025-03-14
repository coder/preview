# Development

Currently, a branch of the coder terraform provider is required.


1. Git clone `git@github.com:coder/terraform-provider-coder.git`
  - Checkout branch `stevenmasley/form_control`
  - Build the provider with `go build -o terraform-provider-coder`
1. Create a file named `.terraformrc` in your `$HOME` directory
1. Add the following content:

```hcl
 provider_installation {
     # Override the coder/coder provider to use your local version
     dev_overrides {
       "coder/coder" = "/path/to/terraform-provider-coder"
     }

     # For all other providers, install them directly from their origin provider
     # registries as normal. If you omit this, Terraform will _only_ use
     # the dev_overrides block, and so no other providers will be available.
     direct {}
 }
```

Now you are using the right terraform provider.