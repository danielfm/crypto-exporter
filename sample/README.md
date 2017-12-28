# Crypto Exporter Dashboard

This is a simple Docker Compose configuration that starts up the exporter,
[Prometheus](https://prometheus.io), and [Grafana](https://grafana.net) for
collecting, storing, and visualizing cryptocurrency trading data exported by
crypto-exporter.

![dashboard](./img/dashboard.png)

## Dependencies

- [Docker](https://docker.com)

## Instructions

Run the following command to boot all services in the required order:

```
$ docker-compose up
```

Then, log into grafana at <http://localhost:3000> with the admin/admin user.
