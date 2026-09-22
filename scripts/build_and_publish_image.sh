#!/usr/bin/env bash

# esse script constrói e publica a imagem docker.
# o script utiliza os seguintes argumentos:
#
# user - usuário para autenticação de registro de container
# token - senha para autenticação de registro de container
# image_name - nome da imagem do container
# image_version - versão da imagem do container

# todo: olhar o drying up: https://gitlab.com/gitlab-org/gitlab-web-ide-vscode-fork/-/issues/7
# ver: https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html
set -o errexit  # aka -e - deixar imediatamente ao ocorrer erros (http://mywiki.wooledge.org/BashFAQ/105)
set -o xtrace   # aka -x - obter "stacktraces" do bash e ver aonde o script falhou
set -o pipefail # falha quando as pipelines contém um erro (http://www.gnu.org/software/bash/manual/html_node/Pipelines.html)

USER=$1
TOKEN=$2
IMAGE_NAME=$3
IMAGE_VERSION=$4

if [ -z "${USER}" ]; then
	echo "user não está definido"

	exit 1
fi

if [ -z "${TOKEN}" ]; then
	echo "token não está definido"

	exit 1
fi

if [ -z "${IMAGE_NAME}" ]; then
	IMAGE_NAME="registry.encore.com/encore/workspaces/encore-workspaces-proxy"
	
	echo "image_name não está definido. utilizando '${IMAGE_NAME}'"
fi

if [ -z "${IMAGE_VERSION}" ]; then
	IMAGE_VERSION="dev-$(TZ=UTC date '+%Y%m%d%H%M%S')"
	
	echo "image_version não está definido. utilizando '${IMAGE_VERSION}'"
fi

docker buildx inspect multi-platform 2>/dev/null | grep '^Driver:[[:space:]]*docker-container' >/dev/null || \
	docker buildx create --name multi-platform --driver docker-container

echo "construindo imagem de container => ${IMAGE_NAME}:${IMAGE_VERSION}"

docker buildx build \
	--builder multi-platform \
	--platform=linux/amd64,linux/arm64 \
	--provenance=false \
	-t "${IMAGE_NAME}:${IMAGE_VERSION}" \
	-f ./Dockerfile \
	.

echo "logando em registro de container => ${CI_REGISTRY}"
docker login -u "${USER}" -p "${TOKEN}" "${CI_REGISTRY}"

echo "publicando imagem de container => ${IMAGE_NAME}:${IMAGE_VERSION}"

docker buildx build \
	--builder multi-platform \
	--platform=linux/amd64,linux/arm64 \
	--provenance=false \
	-t "${IMAGE_NAME}:${IMAGE_VERSION}" \
	-f ./Dockerfile \
	--push \
	.
