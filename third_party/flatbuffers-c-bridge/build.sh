#!/usr/bin/env bash
set -euo pipefail

srcDir=$(dirname "${BASH_SOURCE[0]}")
buildDir=${srcDir}/cmake-build
installDir=${1:-${buildDir}}

buildType=Release

libPrefix=lib
libExt=.a
exeExt=
configArgs="-DCMAKE_BUILD_TYPE=${buildType}"
buildArgs="-- -j"
buildOutputDir=
if [[ "$(uname)" == MINGW* ]] || [[ "$(uname)" == CYGWIN* ]]; then
    exeExt=.exe

    # MinGW build
    configArgs+=' -G "MinGW Makefiles"'

    # MSVC build
    # configArgs+=' -G "Visual Studio 16 2019" -A x64'
    # libPrefix=
    # libExt=.lib
    # buildOutputDir=/${buildType}
    # buildArgs=
    # buildArgs="-- /m"    fails with "error MSB1008: Only one project can be specified."
fi

function build() {
    echo "******** Configuring & building ********"

    srcDirAbsolute="$(pwd)/$srcDir"
    pwd=$(pwd)
    mkdir -p "$buildDir"

    set -x

    # need to use eval because of quotes in configArgs... bash is just wonderful...
    cd "$buildDir"
    eval "cmake \"$srcDirAbsolute\" $configArgs"
    cd $pwd

    # Note: flatbuffers-c-bridge-test implies flatbuffers, flatbuffers-c-bridge and flatbuffers-c-bridge-flatc
    # We don't specify them explicitly to be compatible with MSVC which allows only one target per cmake call...
    cmake --build "$buildDir" --config ${buildType} --target flatbuffers-c-bridge-test ${buildArgs}
    set +x
}

function install() {
    echo "******** Collecting artifacts ********"
    if [[ "${installDir}" != "${buildDir}${buildOutputDir}" ]]; then
        echo "Copying from ${buildDir}${buildOutputDir} to ${installDir}:"
        cp "${buildDir}${buildOutputDir}"/${libPrefix}flatbuffers-c-bridge${libExt} "$installDir"
        cp "${buildDir}${buildOutputDir}"/${libPrefix}flatbuffers-c-bridge-flatc${libExt} "$installDir"
    fi
    cp "${buildDir}"/_deps/flatbuffers-*-build${buildOutputDir}/${libPrefix}flatbuffers${libExt} "$installDir"
    echo "The compiled libraries can be found here:"
    ls -alh "$installDir"/${libPrefix}flatbuffers*${libExt}
}

function test() {
    echo "******** Testing ********"
    (cd "${buildDir}${buildOutputDir}" && ./flatbuffers-c-bridge-test${exeExt})
}

build
test
install
