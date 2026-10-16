#!/usr/bin/env fish
# Installs vex from its GitHub releases, or updates it:
#
#   curl -fsSL https://raw.githubusercontent.com/ixexel661/vex/main/install/install.fish | fish
#
# --uninstall removes vex again, together with its settings and themes:
#
#   curl -fsSL https://raw.githubusercontent.com/ixexel661/vex/main/install/install.fish | fish -c 'source - --uninstall'
#
# VEX_VERSION picks a release, e.g. v1.2.0 (default: the latest one).
# VEX_INSTALL_DIR picks the directory (default: ~/.local/bin).

set -g repo https://github.com/ixexel661/vex
set -g marker "# added by the vex installer"

function fail
    echo "vex: $argv" >&2
    return 1
end

# platform prints the system in the names of the release files, e.g.
# linux-amd64.
function platform
    set -l os
    switch (uname -s)
        case Linux
            set os linux
        case Darwin
            set os darwin
        case '*'
            fail "there is no vex for "(uname -s)
            return
    end
    set -l arch
    switch (uname -m)
        case x86_64 amd64
            set arch amd64
        case aarch64 arm64
            set arch arm64
        case '*'
            fail "there is no vex for "(uname -m)
            return
    end
    # A shell running under Rosetta sees x86_64 on Apple Silicon.
    if test $os = darwin
        set -l translated (sysctl -in sysctl.proc_translated 2>/dev/null)
        if test "$translated" = 1
            set arch arm64
        end
    end
    echo $os-$arch
end

# version prints the tag of the release to install.
function version
    if test -n "$VEX_VERSION"
        echo v(string replace -r '^v' '' -- $VEX_VERSION)
        return
    end
    set -l url (curl -fsSLI -o /dev/null -w '%{url_effective}' $repo/releases/latest)
    or begin
        fail "cannot reach $repo"
        return
    end
    string match -q '*/releases/tag/*' -- $url
    or begin
        fail "$repo has no release yet"
        return
    end
    string replace -r '.*/' '' -- $url
end

function sha256
    if command -q sha256sum
        sha256sum $argv[1] | cut -d ' ' -f 1
    else
        shasum -a 256 $argv[1] | cut -d ' ' -f 1
    end
end

# download puts a release file into file, if its checksum matches.
function download --argument-names file tag asset
    curl -fsSL -o $file $repo/releases/download/$tag/$asset
    or begin
        fail "cannot download $asset of vex $tag"
        return
    end
    set -l sums (curl -fsSL $repo/releases/download/$tag/SHA256SUMS)
    or begin
        fail "cannot download the checksums of vex $tag"
        return
    end
    set -l want (printf '%s\n' $sums | awk -v name=$asset '$2 == name { print $1 }')
    set -l got (sha256 $file)
    if test -z "$want"; or test "$got" != "$want"
        fail "the checksum of $asset does not match, vex is not installed"
    end
end

# remove_download cleans up after a download that did not get installed.
function remove_download --on-event fish_exit
    set -q tmp; and rm -f $tmp
end

function path_file
    echo $__fish_config_dir/conf.d/vex.fish
end

function add_to_path
    set -l file (path_file)
    mkdir -p (dirname $file)
    printf '%s\n' $marker "fish_add_path -g "(string escape -- $dir) >$file
    echo "Added $dir to PATH in $file"
    echo "Open a new terminal to use vex, or run: source $file"
end

function remove_from_path
    set -l file (path_file)
    if grep -qsF $marker $file
        remove $file
    end
end

# warn_if_shadowed tells when another vex comes first in PATH.
function warn_if_shadowed
    set -l found (command -s vex)
    if test -n "$found[1]"; and test "$found[1]" != $dir/vex
        echo "Note: $found[1] comes first in PATH, so vex starts that one"
    end
end

function config_dir
    if test (uname -s) = Darwin
        echo "$HOME/Library/Application Support/vex"
    else if test -n "$XDG_CONFIG_HOME"
        echo $XDG_CONFIG_HOME/vex
    else
        echo $HOME/.config/vex
    end
end

function remove
    if test -e $argv[1]
        rm -rf -- $argv[1]
        echo "Removed $argv[1]"
    end
end

function install_vex
    command -q curl
    or begin
        fail "installing vex needs curl"
        return
    end
    set -l tag (version); or return
    set -l asset vex-(platform); or return
    echo "Downloading vex $tag ($asset)"
    mkdir -p $dir; or return
    set -g tmp (mktemp $dir/.vex.XXXXXX); or return
    download $tmp $tag $asset; or return
    chmod 755 $tmp
    mv $tmp $dir/vex; or return
    echo "Installed vex $tag to $dir/vex"
    if contains -- $dir $PATH
        warn_if_shadowed
    else
        add_to_path
    end
end

function uninstall_vex
    if not test -e $dir/vex
        echo "There is no vex in $dir"
    end
    remove $dir/vex
    remove_from_path
    remove (config_dir)
end

function main
    if test -z "$HOME"
        fail "HOME is not set"
        return
    end
    set -g dir $HOME/.local/bin
    if test -n "$VEX_INSTALL_DIR"
        set dir $VEX_INSTALL_DIR
    end
    switch "$argv"
        case ''
            install_vex
        case --uninstall
            uninstall_vex
        case '*'
            fail "unknown option $argv, the only option is --uninstall"
    end
end

main $argv; or exit 1
