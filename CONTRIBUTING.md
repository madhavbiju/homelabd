# Contributing to homelabd

First, thank you for considering contributing to `homelabd`!

## Development Setup

1. Clone the repository.
2. Ensure you have Go 1.24+ installed.
3. Run `make build` to compile the binary.
4. Run `make test` before submitting any PRs.

## Architecture

`homelabd` is strictly an API-first application. No UI code should exist in this repository. Ensure any new features adhere to the "Generic" and "Secure by default" principles outlined in our core philosophy (e.g. no arbitrary shell execution).

## Pull Requests

1. Create a descriptive branch name.
2. Ensure your commit messages clearly describe the problem and solution.
3. Update `CHANGELOG.md` in your PR if your change affects users.
