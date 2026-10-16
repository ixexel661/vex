#!/usr/bin/env zsh
# Installs vex from its GitHub releases, or updates it:
#
#   curl -fsSL https://raw.githubusercontent.com/ixexel661/vex/main/install/install.zsh | zsh
#
# --uninstall removes vex again, together with its settings and themes:
#
#   curl -fsSL https://raw.githubusercontent.com/ixexel661/vex/main/install/install.zsh | zsh -s -- --uninstall
#
# VEX_VERSION picks a release, e.g. v1.2.0 (default: the latest one).
# VEX_INSTALL_DIR picks the directory (default: ~/.local/bin).
setopt err_exit no_unset pipe_fail

repo=https://github.com/ixexel661/vex
marker="# added by the vex installer"

fail() {
	print -ru2 -- "vex: $*"
	exit 1
}

# platform prints the system in the names of the release files, e.g.
# darwin-arm64.
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
	print -r -- "$os-$arch"
}

# version prints the tag of the release to install.
version() {
	local url
	if [[ -n ${VEX_VERSION:-} ]]; then
		print -r -- "v${VEX_VERSION#v}"
		return
	fi
	url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$repo/releases/latest") || fail "cannot reach $repo"
	[[ $url == */releases/tag/* ]] || fail "$repo has no release yet"
	print -r -- "${url##*/}"
}

sha256() {
	if (( $+commands[sha256sum] )); then
		sha256sum "$1" | cut -d ' ' -f 1
	else
		shasum -a 256 "$1" | cut -d ' ' -f 1
	fi
}

in_path() {
	(( ${path[(Ie)$dir]} ))
}

startup_file() {
	print -r -- "${ZDOTDIR:-$HOME}/.zshrc"
}

add_to_path() {
	local file line
	file=$(startup_file)
	line="export PATH=\"$dir:\$PATH\" $marker"
	if ! grep -qsxF "$line" "$file"; then
		# The line has to start on a line of its own.
		if [[ -s $file && -n $(tail -c 1 "$file") ]]; then
			print >>"$file"
		fi
		print -r -- "$line" >>"$file"
		print -r -- "Added $dir to PATH in $file"
	fi
	print "Open a new terminal to use vex"
}

remove_from_path() {
	local file kept
	file=$(startup_file)
	grep -qsF "$marker" "$file" || return 0
	kept=$(mktemp)
	grep -vF "$marker" "$file" >"$kept" || true
	# Writing through cat keeps a symlinked file a symlink.
	cat "$kept" >"$file"
	rm -f "$kept"
	print -r -- "Removed vex from PATH in $file"
}

# warn_if_shadowed tells when another vex comes first in PATH.
warn_if_shadowed() {
	local found
	found=$(whence -p vex || true)
	if [[ -n $found && $found != "$dir/vex" ]]; then
		print -r -- "Note: $found comes first in PATH, so vex starts that one"
	fi
}

config_dir() {
	if [[ $(uname -s) == Darwin ]]; then
		print -r -- "$HOME/Library/Application Support/vex"
	else
		print -r -- "${XDG_CONFIG_HOME:-$HOME/.config}/vex"
	fi
}

remove() {
	if [[ -e $1 ]]; then
		rm -rf -- "$1"
		print -r -- "Removed $1"
	fi
}

install_vex() {
	local tag asset sums want
	(( $+commands[curl] )) || fail "installing vex needs curl"
	tag=$(version)
	asset=vex-$(platform)
	print -r -- "Downloading vex $tag ($asset)"
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
	print -r -- "Installed vex $tag to $dir/vex"
	if in_path; then
		warn_if_shadowed
	else
		add_to_path
	fi
}

uninstall_vex() {
	if [[ ! -e $dir/vex ]]; then
		print -r -- "There is no vex in $dir"
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
