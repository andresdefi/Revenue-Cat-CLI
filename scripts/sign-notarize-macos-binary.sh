#!/usr/bin/env bash
set -euo pipefail

binary="${1:-}"

if [[ -z "${binary}" ]]; then
  echo "No binary path supplied; skipping macOS signing."
  exit 0
fi

if [[ "${binary}" != *darwin* ]]; then
  echo "Skipping non-macOS binary: ${binary}"
  exit 0
fi

required=(
  APPLE_DEVELOPER_ID_CERT
  APPLE_DEVELOPER_ID_PASS
  APPLE_TEAM_ID
  APPLE_ID
  APPLE_APP_SPECIFIC_PASSWORD
)

missing=0
for name in "${required[@]}"; do
  if [[ -z "${!name:-}" ]]; then
    echo "Apple signing secret ${name} is not set; leaving ${binary} unsigned."
    missing=1
  fi
done

if [[ "${missing}" -ne 0 ]]; then
  exit 0
fi

if [[ ! -f "${binary}" ]]; then
  echo "Binary not found: ${binary}" >&2
  exit 1
fi

identity="$(
  security find-identity -v -p codesigning |
    awk -F '"' '/Developer ID Application/ { print $2; exit }'
)"

if [[ -z "${identity}" ]]; then
  echo "No Developer ID Application identity found in keychain." >&2
  security find-identity -v -p codesigning >&2 || true
  exit 1
fi

echo "Signing ${binary} with ${identity}"
codesign --force --sign "${identity}" --timestamp --options=runtime "${binary}"
codesign --verify --strict --verbose=2 "${binary}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

archive="${tmpdir}/$(basename "${binary}")-notarization.zip"
ditto -c -k --keepParent "${binary}" "${archive}"

echo "Submitting ${binary} for notarization"
xcrun notarytool submit "${archive}" \
  --apple-id "${APPLE_ID}" \
  --password "${APPLE_APP_SPECIFIC_PASSWORD}" \
  --team-id "${APPLE_TEAM_ID}" \
  --wait \
  --timeout 30m

