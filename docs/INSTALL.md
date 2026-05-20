# Installation Guide

## 1. From GitHub Releases (Pre-built Binaries)

For a quick installation, download the pre-built binary for your platform from the [GitHub Releases page](https://github.com/MohamedLamineAllal/MacOSLeanStorage/releases).

### macOS (Homebrew)
If you have [Homebrew](https://brew.sh) installed, you can install `mls` by tapping our repository:

```bash
brew tap MohamedLamineAllal/mls
brew install mls
```

### macOS (Manual)
```bash
# Apple Silicon
curl -sL https://github.com/MohamedLamineAllal/MacOSLeanStorage/releases/latest/download/mls_Darwin_arm64.tar.gz | tar xz -C /usr/local/bin mls

# Intel
curl -sL https://github.com/MohamedLamineAllal/MacOSLeanStorage/releases/latest/download/mls_Darwin_amd64.tar.gz | tar xz -C /usr/local/bin mls
```


### Linux (DEB/RPM)
```bash
# Debian/Ubuntu
wget https://github.com/MohamedLamineAllal/MrLeanStorage/releases/latest/download/mls_linux_amd64.deb
sudo dpkg -i mls_linux_amd64.deb

# Fedora/RedHat
wget https://github.com/MohamedLamineAllal/MrLeanStorage/releases/latest/download/mls_linux_amd64.rpm
sudo rpm -ivh mls_linux_amd64.rpm
```

### Windows (PowerShell)
```powershell
# Download and extract the latest release
Invoke-WebRequest -Uri "https://github.com/MohamedLamineAllal/MrLeanStorage/releases/latest/download/mls_windows_amd64.zip" -OutFile "mls.zip"
Expand-Archive -Path "mls.zip" -DestinationPath "C:\Program Files\mls"
```

## 2. From Source

If you prefer to build from the latest source:

```bash
git clone git@github.com:MohamedLamineAllal/MrLeanStorage.git /tmp/mls-build && \
cd /tmp/mls-build && \
go build -o mls main.go && \
mv mls /usr/local/bin/mls && \
cd /tmp && rm -rf mls-build
```
