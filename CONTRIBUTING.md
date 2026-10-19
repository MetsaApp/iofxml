# Contributing

Thank you for your help. This document tells you how to make a change.

## Rules

- Use only the Go standard library. Open an issue before you add a dependency.
- Keep the fields of a struct in the order of the schema. The writer uses this order.
- Use synthetic data in test files. Do not use the names of real persons.
- Write commit messages and pull request titles as [Conventional Commits](https://www.conventionalcommits.org/), for example `feat: read FinishTime`. The release process reads them.

## Procedure

1. Open an issue that describes the change.
2. Fork the repository and make a branch.
3. Make the change and add a test for it.
4. Run `go test ./...`.
5. Run `golangci-lint run ./...`.
6. If you changed the output of the writer, run `go test -run TestGolden -update .` and examine the diff.
7. Open a pull request that refers to the issue.

## Validate the output against the schema

The tests do not include the IOF schema, because it has no licence statement. You can validate the golden files on your computer:

1. Download `IOF.xsd` from <https://github.com/international-orienteering-federation/datastandard-v3>.
2. Run `xmllint --noout --schema IOF.xsd testdata/golden/*.xml`.
