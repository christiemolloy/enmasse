FROM quay.io/enmasse/fedora-minimal:31
RUN microdnf install gettext python findutils coreutils tar && microdnf clean all

ARG version
ARG maven_version
ARG commit
ENV VERSION=${version} COMMIT=${commit} MAVEN_VERSION=${maven_version}

ADD target/console-init-${maven_version}-dist.tar.gz /

CMD ["/oauth-proxy/bin/init.sh", "/apps/"]
