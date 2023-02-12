FROM node:16

RUN apt update && apt upgrade -y
RUN npm i -g npm@latest
RUN mkdir /app

WORKDIR /app
COPY package.json package-lock.json .
RUN npm i

COPY . /app
ENV NEXT_PUBLIC_API_URL=http://crossword.theodoretliu.com/api
RUN npm run build

ENTRYPOINT npm run start

