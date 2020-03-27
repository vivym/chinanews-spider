FROM scratch

COPY configs /

COPY chinanews-spider /

ENTRYPOINT ["/chinanews-spider", "--config", "chinanews.json"]
