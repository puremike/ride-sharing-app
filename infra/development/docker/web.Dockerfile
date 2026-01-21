FROM node:22-bookworm-slim

WORKDIR /app

COPY web/package*.json ./

RUN npm ci --omit=dev --verbose

COPY web ./

RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]