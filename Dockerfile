FROM alpine:latest as base

FROM scratch

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY netlify-uploader /bin/netlify-uploader

ENTRYPOINT [ "/bin/netlify-uploader" ]