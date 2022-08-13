# Helm chart

## Install the chart

```
helm install --namespace notion2ical --atomic  notion2ical .
```

## Upgrading the chart

```
helm upgrade --namespace notion2ical --atomic notion2ical .
```

## Using locally with latest dev image

A `values.local.yml` file is provided which will point the installation to the latest `dev` image.
