FROM pytorch/torchserve:latest
COPY ./model_archives/* /home/model-server/model-store
COPY dockerd-entrypoint-custom.sh /usr/local/bin/dockerd-entrypoint-custom.sh
ENTRYPOINT ["/usr/local/bin/dockerd-entrypoint-custom.sh"]
CMD ["serve"]
