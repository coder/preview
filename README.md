<!-- markdownlint-disable MD041 -->
<div align="center">
  <a href="https://coder.com#gh-light-mode-only">
    <img src="./.github/assets/images/logo-black.png" alt="Coder Logo Light" style="width: 128px">
  </a>
  <a href="https://coder.com#gh-dark-mode-only">
    <img src="./.github/assets/images/logo-white.png" alt="Coder Logo Dark" style="width: 128px">
  </a>

<h1>
  Workspace Parameters sourced from Terraform
</h1>

<br>
<br>

[Quickstart](#quickstart) | [Docs](https://coder.com/docs) |
[Why Coder](https://coder.com/why) |
[Premium](https://coder.com/pricing#compare-plans)

[![discord](https://img.shields.io/discord/747933592273027093?label=discord)](https://discord.gg/coder)
[![release](https://img.shields.io/github/v/release/coder/preview)](https://github.com/coder/preview/releases/latest)
[![godoc](https://pkg.go.dev/badge/github.com/coder/preview.svg)](https://pkg.go.dev/github.com/coder/preview)
[![Go Report Card](https://goreportcard.com/badge/github.com/coder/preview)](https://goreportcard.com/report/github.com/coder/preview)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9511/badge)](https://www.bestpractices.dev/projects/9511)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/coder/preview/badge)](https://scorecard.dev/viewer/?uri=github.com%2Fcoder%2Fcoder)
[![license](https://img.shields.io/github/license/coder/preview)](./LICENSE)

</div>

This repository contains a component of Coder that handles workspace parameter
management via Terraform. It's responsible for extracting and managing
[workspace parameters](https://coder.com/docs/admin/templates/extending-templates/parameters)
from Terraform configurations, supporting [Coder's](https://coder.com) core
functionality of creating cloud development environments (like EC2 VMs,
Kubernetes Pods, and Docker containers).

The primary repository for Coder is [here](https://github.com/coder/coder).

<!--Should update this with the new cool form options -->
<p align="center">
  <img src="./.github/assets/images/hero-image.png" alt="Coder Hero Image">
</p>

<!-- TODO: Add a usage section that links to coder/coder doc for how to use the `preview` command in coder cli -->

## Support

Do you have a workspace template that has incorrect parameters? Please open
[workspace template behavior issue](https://github.com/coder/preview/issues/new?template=workspace-template-bug-report.md).

For other bugs, feature requests, etc, feel free to
[open an issue](https://github.com/coder/preview/issues/new).

[Join our Discord](https://discord.gg/coder) to provide feedback on in-progress
features and chat with the community using Coder!
