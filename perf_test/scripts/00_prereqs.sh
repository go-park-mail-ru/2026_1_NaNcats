#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

GO_VERSION=${GO_VERSION:-1.26.1}
VEGETA_VERSION=${VEGETA_VERSION:-12.12.0}
MIGRATE_VERSION=${MIGRATE_VERSION:-4.17.0}
ARCH=amd64

if [ ! -x /usr/local/go/bin/go ]; then
  echo ">> installing Go $GO_VERSION"
  curl -fsSLO "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
  sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-${ARCH}.tar.gz"
  rm -f "go${GO_VERSION}.linux-${ARCH}.tar.gz"
fi

if ! command -v vegeta >/dev/null 2>&1; then
  echo ">> installing vegeta $VEGETA_VERSION"
  curl -fsSLO "https://github.com/tsenart/vegeta/releases/download/v${VEGETA_VERSION}/vegeta_${VEGETA_VERSION}_linux_${ARCH}.tar.gz"
  tar xzf "vegeta_${VEGETA_VERSION}_linux_${ARCH}.tar.gz" vegeta && sudo mv vegeta /usr/local/bin/
  rm -f "vegeta_${VEGETA_VERSION}_linux_${ARCH}.tar.gz"
fi

if ! command -v migrate >/dev/null 2>&1; then
  echo ">> installing golang-migrate $MIGRATE_VERSION"
  curl -fsSL "https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VERSION}/migrate.linux-${ARCH}.tar.gz" | tar xz migrate
  sudo mv migrate /usr/local/bin/
fi

# jq — для разбора JSON-ответа логина в 04_provision_owner.sh
command -v jq >/dev/null 2>&1 || { echo ">> installing jq"; sudo apt-get install -y jq; }

echo ">> versions:"; /usr/local/go/bin/go version; vegeta --version | head -1; migrate -version
