FROM nginx:1.13-alpine
MAINTAINER devops@demandbase.com

# Copy over artifacts
COPY ./public /usr/share/nginx/html
COPY ./goldfish.conf /etc/nginx/conf.d/default.conf
