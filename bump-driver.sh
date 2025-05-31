#!/bin/bash
set -euo pipefail

readonly DRIVERS=("mbim" "qmi" "ccid" "at")

get_next_version() {
    local driver="$1"
    local last_tag=$(git tag --list "driver/$driver/v*" | sort -V | tail -n 1)

    if [[ -z "$last_tag" ]]; then
        echo "v0.0.1"
        return
    fi

    local version=$(echo "$last_tag" | sed "s|driver/$driver/v||")
    local major=$(echo "$version" | cut -d. -f1)
    local minor=$(echo "$version" | cut -d. -f2)
    local patch=$(echo "$version" | cut -d. -f3)
    local next_patch=$((patch + 1))

    echo "v$major.$minor.$next_patch"
}

normalize_version() {
    local version="$1"
    if [[ $version != v* ]]; then
        echo "v$version"
    else
        echo "$version"
    fi
}

validate_driver() {
    local driver="$1"
    if [[ ! " ${DRIVERS[*]} " =~ " $driver " ]]; then
        echo "Error: Invalid driver '$driver'. Valid drivers: ${DRIVERS[*]}" >&2
        exit 1
    fi
}

bump_driver_version() {
    local driver="$1"
    local tag="$2"

    echo "Bumping $driver to $tag"

    git tag -as "driver/$driver/$tag" -m "Bump $driver to $tag"
    git push origin "driver/$driver/$tag"

    refresh_go_package "$driver" "$tag"

    echo "Successfully bumped $driver to $tag"
}

refresh_go_package() {
    local driver="$1"
    local version="$2"
    echo "Refreshing go package for $driver"
    local refreshURL="https://proxy.golang.org/github.com/damonto/euicc-go/driver/$driver/@v/$version.info"
    curl -s "$refreshURL" > /dev/null || {
        echo "Failed to refresh go package for $driver at version $version"
        exit 1
    }
    echo "Successfully refreshed go package for $driver at version $version"
}

bump_driver() {
    local driver="$1"
    local tag="${2:-}"

    validate_driver "$driver"

    if [[ -z "$tag" ]]; then
        tag=$(get_next_version "$driver")
    else
        tag=$(normalize_version "$tag")
    fi

    bump_driver_version "$driver" "$tag"
}

bump_drivers() {
    echo "No driver specified, bumping all drivers..."
    for driver in "${DRIVERS[@]}"; do
        echo "Bumping driver: $driver"
        local next_tag=$(get_next_version "$driver")
        bump_driver_version "$driver" "$next_tag"
    done
}

main() {
    local driver="${1:-}"
    local tag="${2:-}"

    if [[ -z "$driver" ]]; then
        bump_drivers
    else
        bump_driver "$driver" "$tag"
    fi
}

main "$@"
