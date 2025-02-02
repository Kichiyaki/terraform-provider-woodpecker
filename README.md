# Terraform Provider for Woodpecker CI [![Actions Status](https://github.com/Kichiyaki/terraform-provider-woodpecker/actions/workflows/ci.yml/badge.svg)](https://github.com/Kichiyaki/terraform-provider-woodpecker/actions/workflows/ci.yml)[![GitHub tag](https://img.shields.io/github/v/tag/Kichiyaki/terraform-provider-woodpecker?label=release)](https://github.com/Kichiyaki/terraform-provider-woodpecker/releases)[![license](https://img.shields.io/github/license/Kichiyaki/terraform-provider-woodpecker.svg)](https://github.com/Kichiyaki/terraform-provider-woodpecker/blob/master/LICENSE)

A Terraform provider used to interact with [Woodpecker CI](https://woodpecker-ci.org/) resources.

## Developing the provider

**Requirements:**

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23
- [Docker](https://docs.docker.com/engine/install/) (for tests)
- [pre-commit](https://pre-commit.com/) (optional, but recommended)
- [Node.js](https://nodejs.org/en) (LTS, only needed for commitlint)
- [direnv](https://direnv.net/) (optional, but recommended)

```shell
# if you have direnv installed
direnv allow

# install git hooks and required tools (terraform-plugin-docs, golangci-lint)
make install

# run tests
go test ./...
```

## Contact

Dawid Wysokiński - [contact@dwysokinski.me](mailto:contact@dwysokinski.me)
