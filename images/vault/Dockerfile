FROM alpine:3.4

WORKDIR /opt/vault

# trust hashicorp gpg key
RUN apk add --update ca-certificates libcap && \
	rm -rf /var/cache/apk/*

ENV VERSION 0.6.2

# download vault and sums and sig
ADD https://releases.hashicorp.com/vault/${VERSION}/vault_${VERSION}_SHA256SUMS /opt/vault/checksums
ADD https://releases.hashicorp.com/vault/${VERSION}/vault_${VERSION}_SHA256SUMS.sig /opt/vault/checksums.sig
ADD https://releases.hashicorp.com/vault/${VERSION}/vault_${VERSION}_linux_amd64.zip /opt/vault/

# verify download
RUN apk add --update gnupg && \
	gpg --recv 51852D87348FFC4C && \
	gpg --verify checksums.sig checksums && \
	apk del --purge gnupg && rm -rf /var/cache/apk/* && \
	grep linux_amd64 checksums > checksum && sha256sum -c checksum

# extract and install
RUN unzip /opt/vault/vault_${VERSION}_linux_amd64.zip && \
	rm /opt/vault/vault_${VERSION}_linux_amd64.zip && \
	mkdir -p /opt/bin && mv /opt/vault/vault /opt/bin/vault && \
	chmod +x /opt/bin/vault && setcap cap_ipc_lock=+ep /opt/bin/vault

ENTRYPOINT [ "/opt/bin/vault" ]
CMD [ "server", "-config", "/etc/vault.hcl" ]
