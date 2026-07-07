#!/usr/bin/env bash
set -euo pipefail

# Pinned tool versions - update both here when upgrading.
PROTOC_GEN_GO_VERSION="v1.36.11"  # google.golang.org/protobuf/cmd/protoc-gen-go
PROTOC_VERSION="35.1"             # protocolbuffers/protobuf (no leading "v")

ROOT="$(git rev-parse --show-toplevel)"
PROTO_DIR="${ROOT}/internal/test/proto"
OUT_DIR="${ROOT}/internal/test/gen"
BIN_DIR="${ROOT}/bin"

mkdir -p "${OUT_DIR}" "${BIN_DIR}"
export PATH="${BIN_DIR}:${PATH}"

# Install protoc-gen-go at the pinned version into BIN_DIR.
GOBIN="${BIN_DIR}" go install \
  "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}"

# Download protoc at the pinned version if the binary is missing or outdated.
install_protoc() {
  local os arch asset tmp
  os="$(uname -s)"
  arch="$(uname -m)"

  case "${os}-${arch}" in
    Linux-x86_64)   asset="protoc-${PROTOC_VERSION}-linux-x86_64.zip"   ;;
    Linux-aarch64)  asset="protoc-${PROTOC_VERSION}-linux-aarch_64.zip" ;;
    Darwin-x86_64)  asset="protoc-${PROTOC_VERSION}-osx-x86_64.zip"     ;;
    Darwin-arm64)   asset="protoc-${PROTOC_VERSION}-osx-aarch_64.zip"   ;;
    *)
      echo "Unsupported platform: ${os}-${arch}" >&2
      exit 1
      ;;
  esac

  tmp="$(mktemp -d)"
  curl -fsSL -o "${tmp}/protoc.zip" \
    "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${asset}"
  # Extract the binary, then extract include/ into bin/ so the layout is:
  #   bin/protoc
  #   bin/include/google/protobuf/*.proto
  # protoc discovers include/ as <exe_dir>/include, so no explicit -I is needed.
  unzip -o -j "${tmp}/protoc.zip" "bin/protoc"  -d "${BIN_DIR}"
  unzip -o    "${tmp}/protoc.zip" "include/*"   -d "${BIN_DIR}"
  chmod +x "${BIN_DIR}/protoc"
  rm -rf "${tmp}"
}

if ! "${BIN_DIR}/protoc" --version 2>/dev/null | grep -q "${PROTOC_VERSION}"; then
  echo "Installing protoc ${PROTOC_VERSION}..."
  install_protoc
fi

protoc \
  -I "${PROTO_DIR}" \
  --go_out="${OUT_DIR}" \
  --go_opt=paths=source_relative \
  "${PROTO_DIR}"/*.proto

echo "Generated Go stubs in ${OUT_DIR}"

git add internal/test/gen/*.pb.go
