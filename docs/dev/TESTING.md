# Testing & Development Guide

This guide outlines the standard procedures for testing changes and verifying distributions for `mls`.

## 1. Unit & Concurrency Testing
We prioritize safety and concurrency correctness. Every change must be verified using the Go race detector.

```bash
# Run all tests with race detection
go test -race ./...
```

## 2. Build Verification
Ensure the project compiles correctly on your local environment before submitting changes.

```bash
# Build the binary
go build -o mls main.go

# Verify binary functionality
./mls --version
```

## 3. Local Homebrew Formula Testing
To test the Homebrew installation process locally without pushing to the remote repository:

1. Generate the distribution artifacts and formula:
   ```bash
   goreleaser release --snapshot --clean
   ```

2. Install the formula locally from the generated file:
   ```bash
   # Replace with the path to the generated formula (usually in dist/formula/mls.rb)
   brew install --build-from-source dist/mls.rb
   ```

3. Verify the installation:
   ```bash
   mls --version
   ```

## 4. CI/CD Validation
The project uses GitHub Actions with GoReleaser.
- **Trigger**: Pushing a tag (e.g., `git tag v0.1.2 && git push origin v0.1.2`) automatically triggers the build, packaging, and publishing pipeline.
- **Monitoring**: Check the "Actions" tab on your [GitHub Repository](https://github.com/MohamedLamineAllal/MrLeanStorage) to monitor workflow status.
- **Permissions**: Ensure `HOMEBREW_TAP_GITHUB_TOKEN` is set in your repository secrets to allow automated Homebrew tap updates.
