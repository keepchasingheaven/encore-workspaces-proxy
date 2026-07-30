#!/usr/bin/env bash

# esse script constrói e publica o helm chart
#
# utilizando os seguintes argumentos:
# user - usuário para autenticação com registro de pacote
# token - senha para autenticação com registro de pacote
# chart_name - nome do chart
# chart_version - versão do chart
#
# todo: cogitar realizer um drying nesse boilerplate:
# https://gitlab.com/gitlab-org/gitlab-web-ide-vscode-fork/-/issues/7
#
# ver mais:
# https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html
set -o errexit  # aka -e - deixa a sessão imediatamente quando ocorre um erro (https://mywiki.wooledge.org/BashFAQ/105)
set -o xtrace   # aka -x - obtém os "stacktraces" de bash e vê aonde o script falha
set -o pipefail # falha quando as pipelines contém um erro (https://www.gnu.org/software/bash/manual/html_node/Pipelines.html)

USER=$1
TOKEN=$2
CHART_NAME=$3
CHART_VERSION=$4

if [ -z "${USER}" ]; then
	echo "usuário não definido"

	exit 1
fi

if [ -z "${TOKEN}" ]; then
	echo "token não definido"

	exit 1
fi

if [ -z "${CHART_NAME}" ]; then
	CHART_NAME="encore-workspaces-proxy"
	
	echo "nome de chart não definido. utilizando '${CHART_NAME}'"
fi

if [ -z "${CHART_VERSION}" ]; then
    CHART_VERSION="dev-$(TZ=UTC date '+%Y%m%d%H%M%S')"
    
	echo "versão de chart não definida. utilizando '${CHART_VERSION}'"
fi

echo "empacotando helm chart"

helm package ./helm

PROJECT_PATH="${CI_PROJECT_PATH:-github.com/keepchasingheaven/encore-workspaces-proxy}"
URL_ENCODED_PROJECT_PATH=$(echo "${PROJECT_PATH}" | jq -Rr @uri)
URL="https://encore.com/api/v4/projects/${URL_ENCODED_PROJECT_PATH}/packages/helm/api/devel/charts"

echo "publicando helm chart => ${CHART_NAME}-${CHART_VERSION}"

curl --fail-with-body --request POST \
	--form "chart=@${CHART_NAME}-${CHART_VERSION}.tgz" \
	--user ${USER}:${TOKEN} \
	"${URL}"
