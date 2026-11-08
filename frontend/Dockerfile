FROM node:22-alpine AS base
WORKDIR /usr/src/app

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY .npmrc* ./

FROM base AS prod-deps
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --prod --frozen-lockfile --ignore-scripts

# Runs on the build machine's own platform even when building for another one:
# its output (build/) is plain JavaScript, the same for every platform. Only
# prod-deps, whose packages may carry native binaries, runs per target platform.
FROM --platform=$BUILDPLATFORM node:22-alpine AS build
WORKDIR /usr/src/app

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY .npmrc* ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile --ignore-scripts && pnpm rebuild esbuild

COPY . .
# The public origin, fixed at build time (adapter-node 6 has no runtime ORIGIN).
# Left empty for the staging/production image: behind Caddy the origin comes
# from PROTOCOL_HEADER/HOST_HEADER per request. Only local.compose.yaml, with
# no proxy in front, passes it.
ARG KIT_PATHS_ORIGIN=
RUN KIT_PATHS_ORIGIN="$KIT_PATHS_ORIGIN" pnpm run build

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

# adapter-node's request body limit, for the whole server. Raised from its
# 512K default for the practice page video upload (50 MB plus the form's
# overhead), which streams through to the backend without being buffered;
# src/hooks.server.ts holds every other route to the old 512K.
ENV BODY_SIZE_LIMIT=60M

COPY --from=prod-deps /usr/src/app/node_modules/ ./node_modules/
COPY --from=build /usr/src/app/build/ ./build/
COPY --from=build /usr/src/app/package.json ./

EXPOSE 3000 3001

CMD ["build/index.js"]
