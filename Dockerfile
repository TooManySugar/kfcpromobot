FROM scratch
COPY build/release/linux_amd64/app /kfc_promo_bot
COPY ca_certs.pem /etc/ssl/certs/
EXPOSE 80/tcp
WORKDIR /app
CMD ["../kfc_promo_bot"]