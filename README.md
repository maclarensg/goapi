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
2. Alternatively you can download the charts (helm-charts.tgz) from the release page and use that instead of cloning this repo.

```
helm upgrade --install goapi helm-chart.tgz \
  --namespace goapi \
  --create-namespace \
  -f helm/goapi/values.yaml 
```
3. Once deployed, it will look something like this
```
➜ kubectl -n goapi get all,ing
NAME                         READY   STATUS    RESTARTS   AGE
pod/goapi-55f6dc6fd-85c2j    1/1     Running   0          19h
pod/redis-67cbbb7766-hnhhx   1/1     Running   0          19h

NAME            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/goapi   ClusterIP   10.43.106.84   <none>        80/TCP     19h
service/redis   ClusterIP   10.43.48.23    <none>        6379/TCP   19h

NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/goapi   1/1     1            1           19h
deployment.apps/redis   1/1     1            1           19h

NAME                               DESIRED   CURRENT   READY   AGE
replicaset.apps/goapi-5f4cbf7c7c   0         0         0       19h
replicaset.apps/goapi-55f6dc6fd    1         1         1       19h
replicaset.apps/redis-67cbbb7766   1         1         1       19h

NAME                              CLASS   HOSTS               ADDRESS        PORTS   AGE
ingress.networking.k8s.io/goapi   nginx   goapi.localdev.me   192.168.5.15   80      19h
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

## Other Notes
1. The helm chart uses k8s' secrets. Recommend to use kubeseal, so that secrets can be store in repo. 
