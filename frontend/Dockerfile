FROM node:20.11.1-alpine AS build
WORKDIR /app

ARG GOOGLE_CLIENT_ID
ARG GOOGLE_CLIENT_SECRET

ENV GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
ENV GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

RUN corepack enable
COPY ./package.json ./pnpm-lock.yaml ./
RUN pnpm install

COPY . .
RUN pnpm run build

FROM node:20.11.1-alpine

WORKDIR /app
RUN adduser -D runner

COPY --from=build /app/build build/
COPY --from=build /app/node_modules node_modules/ 
COPY package.json pnpm-lock.yaml ./

RUN chown -R runner:runner /app/build

EXPOSE 3000 3001
CMD ["node", "build"]
