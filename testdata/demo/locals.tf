locals {
  isAdmin = contains(data.coder_workspace_owner.me.groups, "admin")

  fe_codes = ["PS", "WS"]
  be_codes = ["CL", "GO", "IU", "PY"]
  teams = {
    "frontend" = {
      "display_name" = "Frontend",
      "codes" = local.fe_codes,
      "description" = "The team that works on the frontend.",
      "icon" = "/icon/desktop.svg"
    },
    "backend" = {
      "display_name" = "Backend",
      "codes" = local.be_codes,
      "description" = "The team that works on the backend.",
      "icon" = "/emojis/2699.png",
    },
    "fullstack" = {
      "display_name" = "Fullstack",
      "codes" = concat(local.be_codes, local.fe_codes),
      "description" = "The team that works on both the frontend and backend.",
      "icon" = "/emojis/1f916.png",
    }
  }

  regions = [
    {
      icon  = "/emojis/1f1fa-1f1f8.png"
      name  = "Pittsburgh"
      value = "us-pittsburgh"
    },
    {
      icon  = "/emojis/1f1eb-1f1ee.png"
      name  = "Helsinki"
      value = "eu-helsinki"
    },
    {
      icon  = "/emojis/1f1e6-1f1fa.png"
      name  = "Sydney"
      value = "ap-sydney"
    },
    {
      icon  = "/emojis/1f1e7-1f1f7.png"
      name  = "São Paulo"
      value = "sa-saopaulo"
    },
    {
      icon  = "/emojis/1f1ff-1f1e6.png"
      name  = "Johannesburg"
      value = "za-jnb"
    }
  ]

  region_values = [for region in local.regions : region.value]
  default_regions = tolist(setintersection(data.coder_workspace_owner.me.groups, local.region_values))
  default_region = length(local.default_regions) > 0 ? local.default_regions[0] : local.region_values[0]
}

output "test" {
  value = local.default_region
}