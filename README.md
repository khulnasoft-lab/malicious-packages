[![Scorecard supply-chain security](https://github.com/khulnasoft-labs/infected-packages/actions/workflows/scorecard.yml/badge.svg)](https://github.com/khulnasoft-labs/infected-packages/actions/workflows/scorecard.yml)

# infected-packages

This repository is a collection of reports of malicious packages.

## Quick Start

To validate all OSV reports after making changes:

```sh
make validate
```

## Example OSV Report

```json
{
  "schema_version": "1.5.0",
  "id": "MAL-2024-XXXX",
  "summary": "Malicious code in [package] ([ecosystem])",
  "details": "This package was flagged as malicious ...",
  "affected": [
    {
      "package": {
        "ecosystem": "npm",
        "name": "example-package"
      },
      "versions": [
        "1.2.34"
      ]
    }
  ],
  "credits": [
    {
      "name": "ExampleSource",
      "type": "FINDER",
      "contact": [
        "https://example.com"
      ]
    }
  ],
  "database_specific": {}
}
```

## Documentation

- [Contributing Guide](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)

## CI & Linting

Validation and linting are run automatically on pull requests. Please ensure your contributions pass validation using `make validate` before submitting.
