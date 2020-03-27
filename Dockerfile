FROM scratch

COPY chinanews-spider /

ENTRYPOINT ["/chinanews-spider"]
