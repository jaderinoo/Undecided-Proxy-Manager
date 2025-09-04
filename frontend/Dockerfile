FROM node:20-slim AS builder

WORKDIR /app
RUN corepack enable
COPY package.json ./
RUN yarn install

COPY . .
RUN yarn build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
RUN echo 'server { listen 6070; location / { root /usr/share/nginx/html; index index.html; try_files $uri $uri/ /index.html; } }' > /etc/nginx/conf.d/default.conf
EXPOSE 6070
CMD ["nginx", "-g", "daemon off;"]
