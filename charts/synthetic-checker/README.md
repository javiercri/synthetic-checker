# synthetic-checker

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

A Helm chart for Kubernetes

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| checks | object | `{}` |  |
| configSources.downstreams | list | `[]` |  |
| configSources.syncInterval | string | `"5m"` |  |
| exposeConfig | bool | `false` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"synthetic-checker"` |  |
| image.tag | string | `"latest"` |  |
| imagePullSecrets | list | `[]` |  |
| informer.informOnly | bool | `false` | when set to true, will prevent the checks from being executed in the local instance |
| informer.upstreams | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| k8sLeaderElection | bool | `false` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| nodepinger | bool | `false` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| prometheus.enabled | bool | `true` |  |
| prometheus.endpoint | string | `"metrics"` |  |
| prometheus.operator.enabled | bool | `false` |  |
| prometheus.operator.namespace | string | `"monitoring"` |  |
| prometheus.operator.serviceMonitor.interval | string | `"15s"` |  |
| prometheus.operator.serviceMonitor.scrapeTimeout | string | `"2s"` |  |
| prometheus.port | int | `8080` |  |
| rbacProxy.clientSecret | string | if rbacProxy is enabled, the chart will use the app's SA if this is not set | The name of a kubernetes service account secret |
| rbacProxy.enabled | bool | `false` |  |
| rbacProxy.image.repository | string | `"quay.io/brancz/kube-rbac-proxy"` |  |
| rbacProxy.image.tag | string | `"latest"` |  |
| replicaCount | int | `1` | set replicaCount > 1 and k8sLeaderElection to true to enable leader/follower HA mode |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| service.containerPort | int | `8080` |  |
| service.port | int | `80` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| statusCodes.degraded | int | `200` |  |
| statusCodes.failed | int | `200` |  |
| tolerations | list | `[]` |  |
| watcher.ingresses | bool | `false` |  |
| watcher.namespaces | list | `[]` |  |
| watcher.requiredLabel | string | `""` |  |

