#!/bin/bash

API_SERVER=localhost:4000
COOKIE_JAR=$(mktemp /tmp/cookies.XXXXXX)

cleanup() {
    rm -rf $COOKIE_JAR
}

_curl_rest_raw() {
    curl \
        -v \
        --no-progress-meter \
        -X POST \
        --header 'Content-Type: application/json' \
        "$@"
}


_curl_rest_validate_response() {
    local json=$(_curl_rest_raw "$@") || {
        local status=$?
        echo 1>&2 "curl $* failed: ${status}"
        exit ${status}
    }
    if ! echo "$json" | jq . > /dev/null
    then
        echo 1>&2 "Malformed output from API $*: json=$json"
        return 1
    else
        echo $json
    fi
}

_curl_rest_cookies() {
    _curl_rest_validate_response \
        --cookie "${COOKIE_JAR}" \
        --cookie-jar "${COOKIE_JAR}" \
        "$@"
}

_curl_rest_post() {
    _curl_rest_cookies \
        -X POST \
        --header 'Content-Type: application/json' \
        "$@"
}

_curl_post() {
    local api_path=$1
    local data=${2:-}
    _curl_rest_post --data "$data" http://$API_SERVER/$api_path
}

_curl_rest_cookies() {
    _curl_rest_validate_response \
        --cookie "${COOKIE_JAR}" \
        --cookie-jar "${COOKIE_JAR}" \
        "$@"
}

_curl_rest_get() {
    _curl_rest_cookies \
        -X GET \
        "$@"
}

_curl_get() {
    local api_path=$1
    _curl_rest_get http://$API_SERVER/$api_path
}


# _curl_get "api/v1/health?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m"
# _curl_get "api/v1/sessions?datname=postgres&start=1723438800000&step=5m"

# _curl_get "api/v1/tuples_dml?datname=postgres&start=1723438800000&step=5m"
# _curl_get "api/v1/sessions?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m"
# _curl_get "api/v1/connection_utilization?datname=postgres&start=1723438800000&step=5m"
# _curl_get "api/v1/transactions_in_progress?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m"

# _curl_get "api/v1/tuples_dml?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m"
# _curl_get "api/v1/tuples_reads?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m"
# _curl_get "api/v1/transactions?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m"

# _curl_get "api/v1/health?datname=postgres&dbidentifier=mohammad-dashti-rds-1&step=5s&start=1724796407182"



# _curl_get "api/v1/activity?database_list=postgres&start=0&end=1&step=5000ms&legend=waits&dim=databases&filterdim=hosts&filterdimselected=host1"


# _curl_get "api/v1/all?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m&end=1723438800000&end=1724799089877"
# _curl_get "api/v1/sessions?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m&end=1723438800000&end=1724799089877"
# _curl_get "api/v1/all?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m&end=1723438800000"

# _curl_get "api/v1/sessions?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m&end=1723438800000&end=1724799089877"
# _curl_get "api/v1/all?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m&end=1723438800000"
_curl_get "api/v1/health?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m&end=1723438800000"
# _curl_get "api/v1/health?datname=postgres&dbidentifier=mohammad-dashti-rds-1&start=1723438800000&step=5m"
