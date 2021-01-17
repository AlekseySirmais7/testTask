FROM ubuntu:18.04 AS release
RUN apt-get update
RUN yes | apt-get install software-properties-common

# nginx
RUN \
  add-apt-repository -y ppa:nginx/stable && \
  apt-get update && \
  apt-get install -y nginx && \
  rm -rf /var/lib/apt/lists/* && \
  chown -R www-data:www-data /var/lib/nginx

COPY nginx.conf /etc/nginx/nginx.conf

WORKDIR /etc/nginx
COPY  ./static /etc/nginx/static

COPY  ./dockerPing.sh .

CMD service nginx restart && bash ./dockerPing.sh
