#!/bin/bash
# install-prerequisites.sh
# Installs KIND, kubectl, helm, and docker for the is-it-up-tho project

set -e

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BOLD}====================================================================="
echo "  is-it-up-tho - Prerequisites Installer"
echo "=====================================================================${NC}"
echo ""

# Detect OS
OS="unknown"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    OS="windows"
fi

echo -e "${BOLD}Detected OS: ${GREEN}$OS${NC}"
echo ""

# Check and install KIND
echo -e "${BOLD}Checking KIND...${NC}"
if command -v kind &> /dev/null; then
    echo -e "${GREEN}✓ KIND is already installed: $(kind version)${NC}"
else
    echo -e "${YELLOW}⚠ KIND not found. Installing...${NC}"

    if [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            brew install kind
        else
            echo -e "${YELLOW}Homebrew not found. Installing KIND manually...${NC}"
            curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-darwin-amd64
            chmod +x ./kind
            sudo mv ./kind /usr/local/bin/kind
        fi
    elif [[ "$OS" == "linux" ]]; then
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
        chmod +x ./kind
        sudo mv ./kind /usr/local/bin/kind
    elif [[ "$OS" == "windows" ]]; then
        if command -v choco &> /dev/null; then
            choco install kind -y
        else
            echo -e "${RED}Please install Chocolatey first: https://chocolatey.org/install${NC}"
            echo -e "${YELLOW}Or download KIND manually: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-from-release-binaries${NC}"
            exit 1
        fi
    fi

    if command -v kind &> /dev/null; then
        echo -e "${GREEN}✓ KIND installed successfully: $(kind version)${NC}"
    else
        echo -e "${RED}✗ Failed to install KIND${NC}"
        exit 1
    fi
fi
echo ""

# Check and install kubectl
echo -e "${BOLD}Checking kubectl...${NC}"
if command -v kubectl &> /dev/null; then
    echo -e "${GREEN}✓ kubectl is already installed: $(kubectl version --client --short 2>/dev/null || kubectl version --client)${NC}"
else
    echo -e "${YELLOW}⚠ kubectl not found. Installing...${NC}"

    if [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            brew install kubectl
        else
            curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
            chmod +x kubectl
            sudo mv kubectl /usr/local/bin/
        fi
    elif [[ "$OS" == "linux" ]]; then
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    elif [[ "$OS" == "windows" ]]; then
        if command -v choco &> /dev/null; then
            choco install kubernetes-cli -y
        else
            echo -e "${RED}Please install Chocolatey first: https://chocolatey.org/install${NC}"
            exit 1
        fi
    fi

    if command -v kubectl &> /dev/null; then
        echo -e "${GREEN}✓ kubectl installed successfully${NC}"
    else
        echo -e "${RED}✗ Failed to install kubectl${NC}"
        exit 1
    fi
fi
echo ""

# Check and install Helm
echo -e "${BOLD}Checking Helm...${NC}"
if command -v helm &> /dev/null; then
    echo -e "${GREEN}✓ Helm is already installed: $(helm version --short)${NC}"
else
    echo -e "${YELLOW}⚠ Helm not found. Installing...${NC}"

    if [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            brew install helm
        else
            curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
        fi
    elif [[ "$OS" == "linux" ]]; then
        curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    elif [[ "$OS" == "windows" ]]; then
        if command -v choco &> /dev/null; then
            choco install kubernetes-helm -y
        else
            echo -e "${RED}Please install Chocolatey first: https://chocolatey.org/install${NC}"
            exit 1
        fi
    fi

    if command -v helm &> /dev/null; then
        echo -e "${GREEN}✓ Helm installed successfully: $(helm version --short)${NC}"
    else
        echo -e "${RED}✗ Failed to install Helm${NC}"
        exit 1
    fi
fi
echo ""

# Check Docker
echo -e "${BOLD}Checking Docker...${NC}"
if command -v docker &> /dev/null; then
    if docker info &> /dev/null; then
        echo -e "${GREEN}✓ Docker is installed and running: $(docker version --format '{{.Client.Version}}')${NC}"
    else
        echo -e "${YELLOW}⚠ Docker is installed but not running${NC}"
        echo -e "${YELLOW}Please start Docker Desktop or Docker Engine${NC}"
        if [[ "$OS" == "macos" ]]; then
            echo -e "${YELLOW}Try: open -a Docker${NC}"
        elif [[ "$OS" == "linux" ]]; then
            echo -e "${YELLOW}Try: sudo systemctl start docker${NC}"
        fi
    fi
else
    echo -e "${RED}✗ Docker is not installed${NC}"
    echo ""
    if [[ "$OS" == "macos" ]]; then
        echo -e "${YELLOW}Install Docker Desktop for Mac:${NC}"
        echo "  1. Download from: https://www.docker.com/products/docker-desktop"
        echo "  2. Or use Homebrew: brew install --cask docker"
    elif [[ "$OS" == "linux" ]]; then
        echo -e "${YELLOW}Install Docker:${NC}"
        echo "  curl -fsSL https://get.docker.com -o get-docker.sh"
        echo "  sudo sh get-docker.sh"
        echo "  sudo usermod -aG docker \$USER"
        echo "  # Log out and back in for group changes to take effect"
    elif [[ "$OS" == "windows" ]]; then
        echo -e "${YELLOW}Install Docker Desktop for Windows:${NC}"
        echo "  1. Download from: https://www.docker.com/products/docker-desktop"
        echo "  2. Or use Chocolatey: choco install docker-desktop"
    fi
    echo ""
    echo -e "${RED}Please install Docker and run this script again${NC}"
    exit 1
fi
echo ""

# Summary
echo -e "${BOLD}====================================================================="
echo "  Installation Summary"
echo "=====================================================================${NC}"
echo ""

ALL_INSTALLED=true

if command -v kind &> /dev/null; then
    echo -e "${GREEN}✓ KIND:    $(kind version)${NC}"
else
    echo -e "${RED}✗ KIND:    Not installed${NC}"
    ALL_INSTALLED=false
fi

if command -v kubectl &> /dev/null; then
    echo -e "${GREEN}✓ kubectl: $(kubectl version --client --short 2>/dev/null | head -1 || kubectl version --client | head -1)${NC}"
else
    echo -e "${RED}✗ kubectl: Not installed${NC}"
    ALL_INSTALLED=false
fi

if command -v helm &> /dev/null; then
    echo -e "${GREEN}✓ Helm:    $(helm version --short)${NC}"
else
    echo -e "${RED}✗ Helm:    Not installed${NC}"
    ALL_INSTALLED=false
fi

if command -v docker &> /dev/null && docker info &> /dev/null; then
    echo -e "${GREEN}✓ Docker:  $(docker version --format '{{.Client.Version}}') (running)${NC}"
elif command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠ Docker:  Installed but not running${NC}"
    ALL_INSTALLED=false
else
    echo -e "${RED}✗ Docker:  Not installed${NC}"
    ALL_INSTALLED=false
fi

echo ""

if $ALL_INSTALLED; then
    echo -e "${GREEN}${BOLD}✓ All prerequisites are installed!${NC}"
    echo ""
    echo "Next steps:"
    echo "  make kind-setup    # Create KIND cluster and install operators"
    echo "  make kind-deploy   # Deploy the full health check stack"
    echo "  make status        # Check deployment status"
else
    echo -e "${YELLOW}${BOLD}⚠ Some prerequisites are missing or not running${NC}"
    echo "Please install the missing tools and try again"
    exit 1
fi

echo ""
echo -e "${BOLD}====================================================================${NC}"
