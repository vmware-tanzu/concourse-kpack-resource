ARG base_image
FROM ${base_image}
USER root
RUN mkdir -p /opt/resource \
    && ln -s /layers/org.cloudfoundry.go-mod/app-binary/bin/resource /opt/resource/check \
    && ln -s /layers/org.cloudfoundry.go-mod/app-binary/bin/resource /opt/resource/in \
    && ln -s /layers/org.cloudfoundry.go-mod/app-binary/bin/resource /opt/resource/out
USER cnb