FROM pytorch/torchserve:latest
COPY ./model_archives/cycleganfloraladd.mar /home/model-server/model-store
# ENTRYPOINT [ "curl" ]
# CMD ["--location","--request","POST","http://localhost:8081/models?initial_workers=1&synchronous=true&url=cycleganfloraladd.mar" ]

COPY dockerd-entrypoint-custom.sh /usr/local/bin/dockerd-entrypoint-custom.sh
# USER root
# RUN chmod +x /usr/local/bin/dockerd-entrypoint-custom.sh \
#     && chown -R model-server /home/model-server

ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint-custom.sh"]
CMD ["serve"]
