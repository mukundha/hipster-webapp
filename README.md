# Hipster Shop: Web App 

This is a sample web app using the APIs of [Hipster Shop](https://github.com/GoogleCloudPlatform/microservices-demo)


[Hipster Shop](https://github.com/GoogleCloudPlatform/microservices-demo) is a Cloud-Native Microservices based ecommerce demo

## Installation

### Local

```
    dep ensure --vendor-only
    go install .

    $GOPATH/bin/frontend
```

### Docker

```
    docker build -t frontend .

    docker run -p8080:8080 -e API_BASE_URL=<Hipster-API-URL>
```
