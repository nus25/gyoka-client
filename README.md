# Gyoka editor client
(experimental)
Gyoka Editor(OpenAPI版)のクライアントライブラリ。

# Update Schema
To update the OpenAPI schema to the latest version from GitHub:

```bash
./update-schema.sh
```

This will:
- Download the latest schema from [nus25/gyoka](https://github.com/nus25/gyoka/blob/main/packages/editor/schema/openapi.json)
- Create a backup of the current schema
- Validate the downloaded JSON
- Restore from backup if validation fails

# Build
## Go
To build, run the following command from the repository root:

```bash
cd ./go/generate/ && go generate
```
## Go AT Protocol
- go tool lexgen # Regenerate at protocol lexicon schema `go tool lexgen  --package gyoka --build-file ./lexgen/gyoka.json ./lexicons/`
