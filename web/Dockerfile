FROM oven/bun:1 AS build

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --ignore-scripts

COPY . .
RUN bun run build

FROM node:22-alpine

WORKDIR /app

COPY --from=build /app/.output ./.output

ENV HOST=0.0.0.0 \
    PORT=3000

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
