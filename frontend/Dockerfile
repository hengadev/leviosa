FROM node:20-alpine AS build
WORKDIR /app
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

COPY ./package.json ./pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm run build

FROM node:20-alpine
WORKDIR /app
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

RUN adduser -D sveltekit

COPY --from=build /app/build ./build
# COPY --from=build /app/.env.production .
COPY --from=build /app/package.json /app/pnpm-lock.yaml ./

RUN chown -R sveltekit:sveltekit /app/build

RUN pnpm install --prod --frozen-lockfile && \
    pnpm store prune && \
    rm -rf /root/.local/share/pnpm/store && \
    rm -rf /root/.cache

EXPOSE 3000
CMD ["node", "build"]
