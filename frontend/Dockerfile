FROM node:20-slim AS builder

WORKDIR /app

# Enable corepack for yarn
RUN corepack enable

# Copy package files
COPY package.json yarn.lock ./

# Install dependencies
RUN yarn install --frozen-lockfile

# Copy source code
COPY . .

# Build the application
RUN yarn build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html

# Create nginx configuration template
RUN echo 'server { listen ${PROD_FRONTEND_PORT}; location / { root /usr/share/nginx/html; index index.html; } }' > /etc/nginx/conf.d/default.conf.template

# Remove the default nginx configuration to prevent conflicts
RUN rm -f /etc/nginx/conf.d/default.conf

# Create custom startup script that processes the template
RUN echo '#!/bin/sh' > /start-nginx.sh && \
    echo 'set -e' >> /start-nginx.sh && \
    echo 'envsubst < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf' >> /start-nginx.sh && \
    echo 'exec nginx -g "daemon off;"' >> /start-nginx.sh && \
    chmod +x /start-nginx.sh

EXPOSE 6070
CMD ["/start-nginx.sh"]
