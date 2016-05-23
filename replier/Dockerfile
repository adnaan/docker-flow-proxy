FROM busybox
ARG PORT
EXPOSE $PORT
ADD zoneinfo.tar.gz /
ADD ca-certificates.crt /etc/ssl/certs/
ADD reply /
CMD ["/reply"]
