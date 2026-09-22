#!/usr/bin/env bash

# todo: olhar o drying up: https://gitlab.com/gitlab-org/gitlab-web-ide-vscode-fork/-/issues/7
# ver: https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html
set -o errexit  # aka -e - deixar imediatamente ao ocorrer erros (http://mywiki.wooledge.org/BashFAQ/105)
set -o xtrace   # aka -x - obter "stacktraces" do bash e ver aonde o script falhou
set -o pipefail # falha quando as pipelines contém um erro (http://www.gnu.org/software/bash/manual/html_node/Pipelines.html)

# contexto kube - padronizando para desktop para evitar mishaps
KUBE_CONTEXT="rancher-desktop"

# reempacotar chart workspaces-proxy localmente e construir nova imagem
IMAGE_VERSION="dev-$(TZ=UTC date '+%Y%m%d%H%M%S')"
IMAGE_NAME="registry.encore.com/encore/workspaces/encore-workspaces-proxy:${IMAGE_VERSION}" 
nerdctl --namespace k8s.io  build -t $IMAGE_NAME .
HELM_CHART_VERSION="0.0.1+dev$(TZ=UTC date '+%Y%m%d%H%M%S')"
echo "empacotando chart helm"
rm -rf ./localdev
mkdir -p ./localdev
echo "utilizando ${HELM_CHART_VERSION} como versão de chart helm"
echo "utilizando ${IMAGE_VERSION} como versão de app"
yq e ".appVersion = \"${IMAGE_VERSION}\"" -i ./helm/chart.yaml
yq e ".version = \"${HELM_CHART_VERSION}\"" -i ./helm/chart.yaml
helm package ./helm --destination ./localdev

git checkout -- helm/Chart.yaml

# dumpar configuração utilizada
helm --kube-context "${KUBE_CONTEXT}" get values encore-workspaces-proxy --namespace encore-workspaces > ./localdev/config.yaml

# rodar nova chart com configuração dumpada
helm --kube-context "${KUBE_CONTEXT}" upgrade encore-workspaces-proxy ./localdev/encore-workspaces-proxy-${HELM_CHART_VERSION}.tgz -n encore-workspaces -f ./localdev/config.yaml
