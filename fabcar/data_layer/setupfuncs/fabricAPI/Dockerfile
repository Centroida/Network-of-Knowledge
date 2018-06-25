FROM node:8.9.4

EXPOSE 3000
COPY . /app
WORKDIR /app

RUN cd /app; npm install
CMD ["npm","start"]