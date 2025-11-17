# Prerequisites Installation Guide

This guide will help you install all the required tools for is-it-up-tho.

## Required Tools

- **KIND** (Kubernetes IN Docker) - For local Kubernetes clusters
- **kubectl** - Kubernetes command-line tool
- **Helm 3** - Package manager for Kubernetes
- **Docker** - Container runtime

## Quick Install (Recommended)

### Automated Installation

Run the automated installer script:

```bash
./install-prerequisites.sh
```

Or use the Makefile:

```bash
make install-prereqs
```

This will automatically install missing tools for your operating system.

### Verify Installation

```bash
make check-kind
```

## Manual Installation

### macOS

#### Using Homebrew (Recommended)

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install all tools
brew install kind kubectl helm docker
```

#### Manual Install

```bash
# KIND
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Docker Desktop
# Download from: https://www.docker.com/products/docker-desktop
```

### Linux

```bash
# KIND
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
# Log out and back in for group changes to take effect
```

### Windows

#### Using Chocolatey (Recommended)

```powershell
# Install Chocolatey if not already installed
# See: https://chocolatey.org/install

# Install all tools
choco install kind kubernetes-cli kubernetes-helm docker-desktop -y
```

#### Manual Install

1. **KIND**: Download from [KIND releases](https://github.com/kubernetes-sigs/kind/releases)
2. **kubectl**: Download from [Kubernetes releases](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/)
3. **Helm**: Download from [Helm releases](https://github.com/helm/helm/releases)
4. **Docker Desktop**: Download from [Docker](https://www.docker.com/products/docker-desktop)

## Version Requirements

| Tool | Minimum Version | Recommended |
|------|----------------|-------------|
| KIND | 0.17.0 | 0.20.0+ |
| kubectl | 1.24.0 | Latest stable |
| Helm | 3.10.0 | 3.14.0+ |
| Docker | 20.10.0 | Latest stable |

## Verification

After installation, verify all tools are working:

```bash
# Check versions
kind version
kubectl version --client
helm version
docker --version

# Check Docker is running
docker ps

# Run comprehensive check
make check-kind
```

Expected output:
```
✓ KIND is installed: kind v0.20.0 go1.20.4 darwin/amd64
✓ kubectl is installed: Client Version: v1.28.0
✓ helm is installed: v3.14.0+g3fc9f4b
✓ docker is installed: 24.0.6
```

## Troubleshooting

### KIND Not Found

```bash
# Check if KIND is in PATH
which kind

# If not found, reinstall
./install-prerequisites.sh
```

### kubectl Not Found

```bash
# Verify installation
which kubectl

# If not in PATH, add to .bashrc or .zshrc
export PATH=$PATH:/usr/local/bin
```

### Docker Not Running

#### macOS
```bash
# Start Docker Desktop
open -a Docker

# Wait for Docker to start
until docker ps; do sleep 1; done
```

#### Linux
```bash
# Start Docker service
sudo systemctl start docker

# Enable Docker to start on boot
sudo systemctl enable docker

# Check status
sudo systemctl status docker
```

#### Windows
```powershell
# Start Docker Desktop from Start Menu
# Or from command line
Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe"
```

### Permission Denied (Docker)

#### Linux
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Apply changes (log out and back in, or use)
newgrp docker

# Verify
docker ps
```

### Homebrew Not Found (macOS)

```bash
# Install Homebrew
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Follow the post-install instructions to add Homebrew to PATH
```

### Chocolatey Not Found (Windows)

Open PowerShell as Administrator and run:

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

## Next Steps

Once all prerequisites are installed:

1. **Create KIND cluster:**
   ```bash
   make kind-setup
   ```

2. **Deploy application:**
   ```bash
   make kind-deploy
   ```

3. **Verify deployment:**
   ```bash
   make status
   ```

## Additional Resources

- [KIND Documentation](https://kind.sigs.k8s.io/)
- [kubectl Documentation](https://kubernetes.io/docs/reference/kubectl/)
- [Helm Documentation](https://helm.sh/docs/)
- [Docker Documentation](https://docs.docker.com/)

## Getting Help

If you encounter issues:

1. Run `make check-kind` to verify tool installation
2. Check the [KIND_SETUP.md](KIND_SETUP.md) for detailed setup instructions
3. Review the [Troubleshooting](#troubleshooting) section above
4. Open an issue on GitHub with the error output
