FROM nginx:latest

RUN apt update
RUN apt install vim-tiny -y

COPY ./index.html /usr/share/nginx/html/index.html
COPY ./form.js /usr/share/nginx/html/form.js
COPY ./styles.css /usr/share/nginx/html/styles.css