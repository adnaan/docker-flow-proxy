FROM alpine
RUN apk --update add socat
RUN apk add --update ca-certificates
ARG PORT
EXPOSE $PORT
ADD reply /
ADD entry.sh /
RUN chmod +x /entry.sh
ENTRYPOINT ["/entry.sh"]
CMD ["/entry.sh"]
