FROM node:22-alpine AS base
WORKDIR /usr/src/app

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY .npmrc* ./

FROM base AS prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --frozen-lockfile --ignore-scripts

FROM base AS build
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile --ignore-scripts && pnpm rebuild esbuild

COPY . .
RUN pnpm run build

FROM gcr.io/distroless/nodejs22-debian12:nonroot AS runtime
WORKDIR /usr/src/app

# Commit SHA this image was built from, reported by /healthz so the deploy
# script and the Promote workflow can verify what is actually serving. Pass it
# with: docker build --build-arg COMMIT_SHA=<full-sha> ... Builds without the
# argument (local, ad-hoc) keep the "dev" sentinel instead of reporting a
# misleading version.
#
# This is deliberately the image's only build-time value: everything else the
# app reads (API_URL, COMING_SOON, CLIENT_IP_HEADER, ...) comes from the
# container's runtime environment via $env/dynamic/private, so one image runs
# unchanged with either stack's env file.
ARG COMMIT_SHA=dev
ENV COMMIT_SHA=${COMMIT_SHA}

COPY --from=prod-deps /usr/src/app/node_modules/ ./node_modules/
COPY --from=build /usr/src/app/build/ ./build/
COPY --from=build /usr/src/app/package.json ./

EXPOSE 3000 3001

CMD ["build/index.js"]
