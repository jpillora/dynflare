FROM scratch
COPY dynflare /dynflare
ENTRYPOINT ["/dynflare"]
