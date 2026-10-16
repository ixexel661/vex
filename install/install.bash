#!/usr/bin/env bash
# Installs vex from its GitHub releases, or updates it:
#
#   curl -fsSL https://raw.githubusercontent.com/ixexel661/vex/main/install/install.bash | bash
#
# --uninstall removes vex again, together with its settings and themes:
#
#   curl -fsSL https://raw.githubusercontent.com/ixexel661/vex/main/install/install.bash | bash -s -- --uninstall
#
# VEX_VERSION picks a release, e.g. v1.2.0 (default: the latest one).
# VEX_INSTALL_DIR picks the directory (default: ~/.local/bin).
set -euo pipefail

repo=https://github.com/ixexel661/vex
marker="# added by the vex installer"

fail() {
	echo "vex: $*" >&2
	exit 1
}

# platform prints the system in the names of the release files, e.g.
# linux-amd64.
platform() {
	local os arch
	case $(uname -s) in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*) fail "there is no vex for $(uname -s)" ;;
	esac
	case $(uname -m) in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) fail "there is no vex for $(uname -m)" ;;
	esac
	# A shell running under Rosetta sees x86_64 on Apple Silicon.
	if [[ $os == darwin && $(sysctl -in sysctl.proc_translated 2>/dev/null) == 1 ]]; then
		arch=arm64
	fi
	echo "$os-$arch"
}

# version prints the tag of the release to install.
version() {
	local url
	if [[ -n ${VEX_VERSION:-} ]]; then
		echo "v${VEX_VERSION#v}"
		return
	fi
	url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$repo/releases/latest") || fail "cannot reach $repo"
	[[ $url == */releases/tag/* ]] || fail "$repo has no release yet"
	echo "${url##*/}"
}

sha256() {
	if command -v sha256sum >/dev/null; then
		sha256sum "$1" | cut -d ' ' -f 1
	else
		shasum -a 256 "$1" | cut -d ' ' -f 1
	fi
}

in_path() {
	[[ ":$PATH:" == *":$dir:"* ]]
}

# startup_file is the file bash reads when a new terminal opens. On macOS,
# that is a login shell, which only reads the first of its files that exists.
startup_file() {
	local file
	if [[ $(uname -s) != Darwin ]]; then
		echo "$HOME/.bashrc"
		return
	fi
	for file in .bash_profile .bash_login .profile; do
		if [[ -e $HOME/$file ]]; then
			echo "$HOME/$file"
			return
		fi
	done
	echo "$HOME/.bash_profile"
}

add_to_path() {
	local file line
	file=$(startup_file)
	line="export PATH=\"$dir:\$PATH\" $marker"
	if ! grep -qsxF "$line" "$file"; then
		# The line has to start on a line of its own.
		if [[ -s $file && -n $(tail -c 1 "$file") ]]; then
			echo >>"$file"
		fi
		echo "$line" >>"$file"
		echo "Added $dir to PATH in $file"
	fi
	echo "Open a new terminal to use vex"
}

remove_from_path() {
	local file kept
	for file in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.bash_login" "$HOME/.profile"; do
		grep -qsF "$marker" "$file" || continue
		kept=$(mktemp)
		grep -vF "$marker" "$file" >"$kept" || true
		# Writing through cat keeps a symlinked file a symlink.
		cat "$kept" >"$file"
		rm -f "$kept"
		echo "Removed vex from PATH in $file"
	done
}

# warn_if_shadowed tells when another vex comes first in PATH.
warn_if_shadowed() {
	local found
	found=$(command -v vex || true)
	if [[ -n $found && $found != "$dir/vex" ]]; then
		echo "Note: $found comes first in PATH, so vex starts that one"
	fi
}

config_dir() {
	if [[ $(uname -s) == Darwin ]]; then
		echo "$HOME/Library/Application Support/vex"
	else
		echo "${XDG_CONFIG_HOME:-$HOME/.config}/vex"
	fi
}

remove() {
	if [[ -e $1 ]]; then
		rm -rf -- "$1"
		echo "Removed $1"
	fi
}

install_vex() {
	local tag asset sums want
	command -v curl >/dev/null || fail "installing vex needs curl"
	tag=$(version)
	asset=vex-$(platform)
	echo "Downloading vex $tag ($asset)"
	mkdir -p "$dir"
	tmp=$(mktemp "$dir/.vex.XXXXXX")
	trap 'rm -f "$tmp"' EXIT
	curl -fsSL -o "$tmp" "$repo/releases/download/$tag/$asset" || fail "cannot download $asset of vex $tag"
	sums=$(curl -fsSL "$repo/releases/download/$tag/SHA256SUMS") || fail "cannot download the checksums of vex $tag"
	want=$(awk -v name="$asset" '$2 == name { print $1 }' <<<"$sums")
	if [[ -z $want || $(sha256 "$tmp") != "$want" ]]; then
		fail "the checksum of $asset does not match, vex is not installed"
	fi
	chmod 755 "$tmp"
	mv "$tmp" "$dir/vex"
	echo "Installed vex $tag to $dir/vex"
	if in_path; then
		warn_if_shadowed
	else
		add_to_path
	fi
}

uninstall_vex() {
	if [[ ! -e $dir/vex ]]; then
		echo "There is no vex in $dir"
	fi
	remove "$dir/vex"
	remove_from_path
	remove "$(config_dir)"
}

main() {
	[[ -n ${HOME:-} ]] || fail "HOME is not set"
	dir=${VEX_INSTALL_DIR:-$HOME/.local/bin}
	case "$*" in
	"") install_vex ;;
	--uninstall) uninstall_vex ;;
	*) fail "unknown option $*, the only option is --uninstall" ;;
	esac
}

main "$@"
