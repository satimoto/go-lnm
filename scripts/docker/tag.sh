#!/bin/sh
COMMIT_HASH=$(git log -1 --pretty=%h)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
VERSION=$(echo ${BRANCH} | sed -e 's/.*\/v//g' | sed -e 's/\/.*//g')
UNCOMMITED_FILES=$(git status -s | wc -l | tr -d ' ')

usage()
{
    echo ""
    echo "Usage: tag [OPTIONS]"
    echo ""
    echo "Tags the docker image with the git branch version and git commit hash (e.g. 1.12.0-4e543431)"
    echo ""
    echo "Options:"
    echo "  -b, --build    Rebuild the docker image"
    echo "  -t, --testnet  Tag the image for testnet"
    echo "  -m, --mainnet  Tag the image for mainnet"
    echo "  -c, --commit   Push all tagged images"
    echo "  -f, --force    Forces continuation if there are uncommitted files, otherwise user is prompted"
    echo ""
}

while [ $# -ge 1 ]; do
    case $1 in
        -b|--build) BUILD="SET"            
        ;;
        -t|--testnet) TESTNET="SET"            
        ;;
        -m|--mainnet) MAINNET="SET"            
        ;;
        -c|--commit) COMMIT="SET"            
        ;;
        -f|--force) FORCE="SET"            
        ;;
        -h|--help) usage
                   exit
                   ;;
        *) usage
           exit 1
    esac
    shift
done

if [ "$UNCOMMITED_FILES" -ne "0" ] && [ -z "$FORCE" ]; then
    read -p "There are ${UNCOMMITED_FILES} uncommitted file(s). Continue? [Y/n]: " CONFIRM

    if [ "$CONFIRM" != "Y" ]; then
        exit 1
    fi
fi

if [ -z "$VERSION" ]; then
    VERSION=${BRANCH}
fi

if [ -n "$BUILD" ]; then
    echo "Building lsp"
    docker build -t lsp .
fi

if [ -n "$TESTNET" ]; then
    docker tag lsp 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:testnet
    echo "Tagged 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:testnet"
fi

docker tag lsp 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}
echo "Tagged 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}"
docker tag lsp 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}-${COMMIT_HASH}
echo "Tagged 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}-${COMMIT_HASH}"

if [ -n "$MAINNET" ]; then
    docker tag lsp 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:mainnet
    echo "Tagged 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:mainnet"
fi

if [ -n "$COMMIT" ]; then
    if [ -n "$TESTNET" ]; then
        docker push 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:testnet
        echo "Pushed 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:testnet"
    fi

    docker push 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}
    echo "Pushed 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}"
    docker push 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}-${COMMIT_HASH}
    echo "Pushed 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:${VERSION}-${COMMIT_HASH}"

    if [ -n "$MAINNET" ]; then
        docker push 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:mainnet
        echo "Pushed 526438337184.dkr.ecr.eu-west-1.amazonaws.com/lsp:mainnet"
    fi
fi