# Release Process & Best Practices

To maintain high quality and reliability for `mls`, follow these release best practices.

## 1. Semantic Versioning (SemVer)
We follow [Semantic Versioning 2.0.0](https://semver.org/).
- **MAJOR**: Incompatible API changes or major architectural refactors.
- **MINOR**: Backward-compatible functionality additions.
- **PATCH**: Backward-compatible bug fixes or performance improvements.

## 2. Release Workflow
1.  **Develop**: Perform work on a specific feature or fix branch.
2.  **Validate**: Ensure all tests pass (`go test -race ./...`) and the build is stable.
3.  **CHANGELOG**: Document all notable changes in `CHANGELOG.md` (create it if it doesn't exist).
4.  **Tagging**: Use the following format for git tags: `vX.Y.Z`.
    ```bash
    git tag -a v0.1.0 -m "Release description"
    git push origin v0.1.0
    ```
5.  **CI/CD**: Pushing a `v*` tag triggers the GitHub Actions workflow defined in `.github/workflows/release.yml`, which compiles and uploads binaries for all platforms.

## 3. Best Practices
- **Never mutate release tags**: If a mistake is found, release a new version (e.g., `v0.1.1`).
- **Clean Main**: Keep the `main` branch always in a releasable state.
- **Documentation**: Ensure `README.md` and `USER_GUIDE.md` are updated before creating a tag.
- **Automated Builds**: Always rely on the CI/CD pipeline for binary distribution to ensure consistency across platforms (macOS, Linux, Windows).
