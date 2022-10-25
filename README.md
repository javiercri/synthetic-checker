# Synthetic Checker

The synthetic checker is a tool to run synthetic checks and track their statusses.
For now the tool supports HTTP, DNS and Kubernetes checks but more types of checks can be added in the future.
Checks are executed periodically and asynchronously in the background.

To make development easier, a [Makefile](./Makefile) is provided, run `make` with no arguments to get a list of all available options.

To run all the linters, tests and build a binary, run `make build`

## Usage

Check the `--help` flag to see all available options:

```console
$ go run main.go --help
A service to run synthetic checks and report their results.

Usage:
  synthetic-checker [flags]

Flags:
  -C, --certFile string            File containing the x509 Certificate for HTTPS.
  -c, --config string              config file (default is $HOME/.checks.yaml)
  -d, --debug                      Set log level to debug
  -D, --degraded-status-code int   HTTP status code to return when check check is failed (default 200)
  -F, --failed-status-code int     HTTP status code to return when all checks are failed (default 200)
  -h, --help                       help for synthetic-checker
      --k8s-leader-election        Enable leader election, only works when running in k8s
  -K, --keyFile string             File containing the x509 private key for HTTPS.
  -P, --pass string                Set BasicAuth password for the http listener
  -p, --port int                   Port for the http listener (default 8080)
      --pretty-json                Pretty print JSON responses
  -l, --request-limit int          Max requests per second per client allowed
  -s, --securePort int             Port for the HTTPS listener (default 8443)
  -S, --strip-slashes              Strip trailing slashes befofore matching routes
  -U, --user string                Set BasicAuth user for the http listener
```

## Configuration

By default the tool will look for a configuration file in one of the following locations:

- The location from where the tool was started: `./checks.yaml`
- The current user's home directory: `${HOME}/checks.yaml`
- The `/etcd` directory: `/etc/config/checks.yaml`

A sample configuration file can be found in [checks.yaml](./checks.yaml)

```yaml
---
httpChecks:
  stat503:
    url: https://httpstat.us/503
    interval: 10s
  stat200:
    url: https://httpstat.us/200
    interval: 10s
    initialDelay: 2s
dnsChecks:
  google:
    host: "www.google.com"
    interval: 15s
k8sChecks:
  coredns:
    kind: "Deployment.v1.apps"
    name: "coredns"
    namespace: "kube-system"
    interval: 20s
```

## Running the service

You can run the service locally with docker or using the built binary directly.

### Using go run

The simplest way to start the service is through the `go run` command:

```sh
go run main.go
```

Once the service has been started, you can try its endpoints with curl.

Check the global status of all checks:

```console
$ curl -s http://localhost:8080/ | jq .
{
  "stat200-http": {
    "ok": true,
    "timestamp": "2022-08-07T15:30:45.035296+01:00",
    "duration": 438158210,
    "contiguousFailures": 0,
    "timeOfFirstFailure": "0001-01-01T00:00:00Z"
  },
  "stat503-http": {
    "error": "Unexpected status code: '503' expected: '200'",
    "timestamp": "2022-08-07T15:30:50.033884+01:00",
    "duration": 420079871,
    "contiguousFailures": 14,
    "timeOfFirstFailure": "2022-08-07T15:28:40.045752+01:00"
  }
}
```

Or check the metrics endpoint:

```console
$ curl -s http://localhost:8080/metrics | grep 'check_status_up'
# HELP check_status_up Status from the check
# TYPE check_status_up gauge
check_status_up{name="stat200"} 1
check_status_up{name="stat503"} 0
```

### Using Docker

You can run the service through Docker with the following commands:

```sh
docker build --rm --tag synthetic-checker
docker run -it --mount type=bind,source=${PWD}/checks.yaml,target=/etc/config/checks.yaml -p 0.0.0.0:8080:8080 synthetic-checker
```

note: You can also build the docker image using the provided `Makefile` with `make docker-build`

### In Kubernetes using Helm

You can run the service in a Kubernetes cluster using the provided Dockerfile and helm chart.

Build and publish the docker image:

```sh
make docker-build
make docker-release # add -e DOCKER_REGISTRY="my.reg.com" to override Docker registry
```

Create a helm values file with your configuration overrides, you can see an example in [helm/synthetic-checker/ci/with_checks.yaml](./helm/synthetic-checker/ci/with_checks.yaml),
Or have a look at  the default [values](./helm/synthetic-checker/values.yaml) to see all the available options.

And deploy the service using the following command:

```sh
helm upgrade --install -n <target_namespace> -f <path/to/your/custom_values.yaml> synthetic-checker ./helm/synthetic-checker
```

#### HA modes

When running in Kubernetes, you have 2 options for running in HA mode.

- Running multiple independent instances, where each will execute its own checks
  To use this mode set the `replicaCount` to any number higher than 1
- Running multiple instances with leader election, were the leader will execute the checks and the followers will sync the results from it.
  To use this mode set the `replicaCount` to any number higher than 1 and `k8sLeaderElection` to `true`

### In local Kubernetes using colima

If you don't have `colima` installed, have a look at [colima's GitHub page](https://github.com/abiosoft/colima)

Start colima with Kubernetes enabled:

```sh
colima start --kubernetes
```

Once colima is ready, build the docker image with it

```sh
make docker-build
```

Once the image is built, colima makes it available in the Kubernetes cluster and you can install the service using helm:

```sh
helm upgrade --install -f helm/synthetic-checker/ci/with_checks.yaml synthetic-checker ./helm/synthetic-checker
```

To test the service, you can use kubectl port forwarding:

```sh
kubectl port-forward svc/synthetic-checker 8080:80
```

And from another terminal window, tab or Tmux pane, you can reach the service with:

```sh
curl -s http://localhost:8080/
```

## Monitoring synthetic checkes with Prometheus

Install Prometheus:

```sh
kubectl create namespace monitoring
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --namespace monitoring --install kube-stack-prometheus prometheus-community/kube-prometheus-stack --set prometheus-node-exporter.hostRootFsMount.enabled=false --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false --set prometheus.prometheusSpec.probeSelectorNilUsesHelmValues=false --set prometheus.prometheusSpec.ruleSelectorNilUsesHelmValues=false --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false --set thanosRuler.thanosRulerSpec.ruleSelectorNilUsesHelmValues=false
```

Update the synthetic-checker installation to include the Prometheus operator related resources:

```sh
helm upgrade --install -f helm/synthetic-checker/ci/with_checks.yaml -f helm/synthetic-checker/ci/with_prom_op.yaml synthetic-checker ./helm/synthetic-checker
```

Connecting to the Prometheus web UI:

```sh
kubectl port-forward --namespace monitoring svc/kube-stack-prometheus-kube-prometheus 9090:9090
```

Opening a browser tab on http://localhost:9090 shows the Prometheus web UI, the following URL will give you the status of your HTTP checks:

```text
http://localhost:9090/graph?g0.expr=check_status_up&g0.tab=0&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h
```

Connecting To Grafana:

The credentials to connect to the Grafana web interface are stored in a Kubernetes Secret and encoded in base64. We retrieve the username/password couple with these two commands:

```sh
kubectl get secret --namespace monitoring kube-stack-prometheus-grafana -o jsonpath='{.data.admin-user}' | base64 -d
kubectl get secret --namespace monitoring kube-stack-prometheus-grafana -o jsonpath='{.data.admin-password}' | base64 -d
```

We create the port-forward to Grafana with the following command:

```sh
kubectl port-forward --namespace monitoring svc/kube-stack-prometheus-grafana 8080:80
```

Open your browser and go to http://localhost:8080 and fill in previous credentials

A dashboard is included in the provided helm chart, search for "synthetic checks" in the Grafana UI.

![dashboard](./imgs/dash.png)
