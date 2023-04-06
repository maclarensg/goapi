# GOAPI

A challenge to create a RESTAPI API service.

The service is bundle with 2 components
1) The goapi server 
2) A redis that acts as its query store


## TL;DR

### To run it locally

1. Run `docker compose up -d`

### To Check the support api documentation

1. Open your browser and navigate to `http://localhost:3000/swagger/index.html`

### To deploy to a k8s cluster 

1. Run the following command to deploy

```
helm upgrade --install goapi ./charts/goapi \
  --namespace goapi \ 
  --create-namespace \
  -f helm/goapi/values.yaml
```

### To run a local test as you code

1. Run `make test`. This will target the 'test' target of the Dockerfile which perform simple go testing. It's compliant with rancher desktop, so it will replace `docker` with `nerdctl`.


## Demo

## Pre-requisite
1. Install Rancher Desktop on your machine
2. Deploy the helm chart to your k3s running on your machine

## Accessing the Application
1. You can reach out to the application by entering `http://goapi.localdev.me`, or
2. curl
```
➜  ~ curl goapi.localdev.me
{"version":"0.1.0","date":1680761515,"kubernetes":true}%
```

