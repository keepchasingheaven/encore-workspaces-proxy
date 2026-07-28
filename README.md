# encore-workspaces-proxy

com o `encore-workspaces-proxy`, você pode autenticar e autorizar os [workspaces](https://docs.encore.com/user/workspace/) rodando no cluster. o `encore-workspaces-proxy` utiliza um design de proxy central e automaticamente descobre backends baseado em anotações no serviço kubernetes.

## processo de lançamento

- atualiza a versão da imagem do container caso haja qualquer mudança no código da aplicação ou helm chart. o valor será a nova tag git que foi pushada.
	- `appversion` em `helm/chart.yaml`.
- atualiza a versão do helm chart caso haja qualquer mudança no código da aplicação ou helm chart. o valor será diferente da tag git que foi pushada.
	- `version` em `helm/chart.yaml`.
- atualiza a versão do helm chart nas [documentações de desenvolvimento local](https://github.com/keepchasingheaven/encore-workspaces-proxy/README.md)
- mescla as mudanças acima para a branch `main`.
- cria e pusha uma nova tag git da branch `main`.
	```
	# atualiza a tag_name
	tag_name=0.0
	
	git tag -a "${tag_name}" -m "versão ${tag_name}"
	git push origin "${tag_name}"
	```

## rbac

o proxy de workspaces do encore suporta diversos métodos de rbac:

### 1. todos os namespaces

este é o método padrão. a chart irá criar um `clusterrole` e `clusterrolebinding` que dá a permissão de encontrar workspaces em todos os namespaces.

### 2. namespaces específicos

a chart irá criar um `rolebinding` e `role` em cada um dos namespaces listados.

```bash
helm install helm/ --set="{workspaces-1,workspaces-2}"
```

### 3. controlado por usuário

a chart não criará nenhum recurso rbac e você precisará implementá-lo por conta própria.
use isso quando o dispositivo ou a pessoa que instala o gráfico não tiver acesso aos namespaces de destino no momento da implantação.
