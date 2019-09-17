ARG GOLANG_BASE=alpine
ARG GOLANG_BASE_DIGEST=

FROM $GOLANG_BASE$GOLANG_BASE_DIGEST

RUN mkdir /bcbin/
WORKDIR /bcbin/

ADD s3 /usr/bin/
ADD cloudformation /usr/bin/

RUN chmod +x /usr/bin/s3
RUN chmod +x /usr/bin/cloudformation
