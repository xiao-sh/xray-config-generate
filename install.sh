REPO="xiao-sh/xray-config-generate"

OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
Linux) GOOS="linux" ;;
Darwin) GOOS="darwin" ;;
*)
	echo "Unsupported OS"
	exit 1
	;;
esac

case "$ARCH" in
x86_64) GOARCH="amd64" ;;
arm64 | aarch64) GOARCH="arm64" ;;
*)
	echo "Unsupported ARCH"
	exit 1
	;;
esac

FILENAME="myapp-${GOOS}-${GOARCH}"

URL="https://github.com/${REPO}/releases/latest/download/${FILENAME}"

TMP_FILE="/tmp/${FILENAME}"

curl -fL "$URL" -o "$TMP_FILE"
chmod +x "$TMP_FILE"

exec "$TMP_FILE" "$@"
