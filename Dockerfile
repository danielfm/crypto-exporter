FROM centurylink/ca-certs
MAINTAINER Daniel Martins <daniel.tritone@gmail.com>

COPY ./bin/crypto-exporter /crypto-exporter
ENTRYPOINT ["/crypto-exporter"]
