# SPDX-License-Identifier: Apache-2.0

FROM postgres:16.3 as base

RUN addgroup --system autodba && adduser --system --group autodba --home /home/autodba --shell /bin/bash

# set environment variables
ENV PYTHONDONTWRITEBYTECODE 1
#ENV PYTHONUNBUFFERED 1  # this works around setups where line buffering is disabled; it should not be needed here

RUN apt-get update
RUN apt-get install -y --no-install-recommends \
    nodejs          \
    npm             \
    procps          \
    python3         \
    python3-venv    \
    python3-pip

USER autodba
WORKDIR /home/autodba
RUN mkdir src elm

# Create a Python virtual environment
RUN python3 -m venv /home/autodba/venv

# Activate virtual environment
ENV PATH="/home/autodba/venv/bin:$PATH"

# install + cache python dependencies
WORKDIR /home/autodba/src
COPY src/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM base as builder

# install + cache npm dependencies
WORKDIR /home/autodba/elm
COPY --chown=autodba:autodba elm/package.json .
RUN npm install

# install + cache python dependencies
WORKDIR /home/autodba/elm
COPY --chown=autodba:autodba elm .
RUN npm run build  # creates /home/autodba/elm/dist_prod

WORKDIR /home/autodba/src
COPY --chown=autodba:autodba src .
# If we decide to ship compiled python, the command that does it belongs here.

FROM builder as lint
WORKDIR /home/autodba/src
RUN flake8 --ignore=E501,F401,E302,E305 .

WORKDIR /home/autodba/elm
RUN npm run check-format

FROM builder as test

WORKDIR /home/autodba/elm
RUN npm run format # todo...

# TODO: pytest generates a covearge report that we lose.  Write documentation
#       (or a script) that bind mounts the source checkout on top of . and runs
#       `python -m pytest` in the builder container.
WORKDIR /home/autodba/src
RUN POSTGRES_DB=phony_db AUTODBA_TARGET_DB=postgresql://phony_db_user:phony_db_pass@localhost:5432/phony_db python -m pytest

FROM base as autodba

USER root
COPY --from=builder /home/autodba/src /home/autodba/src
COPY --from=builder /home/autodba/elm/dist_prod /home/autodba/elm/dist_prod

WORKDIR /home/autodba/src

# Expose port 8080 for HTTP traffic
EXPOSE 8080

# TODO: Later, when we support external agent DBs (maybe never?) set these using docker args.
ENV FLASK_APP=api/endpoints.py
# may as well hardcode this on until we have a runtime configuration that we're comfortable with.
# Flask makes it clear that our current setup shouldn't be used in production.
ENV FLASK_DEBUG=True
ENV POSTGRES_DB=autodba_db
ENV POSTGRES_USER=autodba_db_user
ENV POSTGRES_PASSWORD=autodba_db_pass
ENV POSTGRES_HOST=localhost
ENV POSTGRES_PORT=5432

# run entrypoint.sh
CMD ["./entrypoint.sh"]
