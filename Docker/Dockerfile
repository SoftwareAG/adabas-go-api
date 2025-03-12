FROM store/softwareag/adabas-ce:6.7.0 AS adabas

VOLUME /opt


FROM golang:1.12

COPY --from=adabas /opt/softwareag/bin /opt/softwareag/bin
COPY --from=adabas /opt/softwareag/common /opt/softwareag/common
COPY --from=adabas /opt/softwareag/AdabasClient /opt/softwareag/AdabasClient

# User id and group id used for the sagadmin user
ENV sag_user_id 1724
ENV group_id 1724

RUN groupadd -g $group_id sagadmin && \
    useradd -d /opt/softwareag -m -s /bin/bash -u $sag_user_id -g sagadmin sagadmin

COPY --from=adabas /etc/profile.d/sagenv.sh /etc/profile.d/sagenv.sh
COPY goenv.sh /etc/profile.d/goenv.sh

# Software GmbH basic environment
ENV SAG /opt/softwareag
ENV ADAPROGDIR /opt/softwareag/Adabas

# Remove old scripts and create environment
RUN cd /opt/softwareag && \
    mkdir /data && \
    chmod 775 /data && \
    chmod 644 /etc/profile.d/goenv.sh && \
    chmod 775 /opt/softwareag


USER sagadmin
ENV HOME /data
ENV SAG_HOME /opt/softwareag
WORKDIR /data

VOLUME /data
CMD ["bash","-l"]

