# Prometheus Configuration

## Default `query-max-concurrency` Setting: **4**

The `query-max-concurrency` setting in Prometheus controls the maximum number of queries that can be executed concurrently.
Setting this value appropriately is crucial to ensure that Prometheus can handle the load, as system instability can occur if the memory resources are exceeded

### Why Set `query-max-concurrency` to **4**?

A concurrency setting of **4** is a reasonable starting point for most environments. As a rule of thumb, the number of cpu's avaialable should be used. There isn't any reason to go higher unless the workload is IO-bound. Here's why:
Limiting concurrency helps avoid excessive CPU or memory usage, especially during periods of high query traffic. Prometheus will still be able to handle other tasks (like data ingestion and internal processing) without degrading system performance.

### What Happens if the Concurrency Limit is Exceeded?

If the number of concurrent queries exceeds the `query-max-concurrency` setting , Prometheus will queue excess queries until a slot becomes available. Here's what happens:

1. **Queuing of Queries:**
   - Queries beyond the concurrency limit will be placed in a queue. They will wait until one of the currently running queries completes and frees up a slot.

2. **Increased Latency:**
   - When queries are queued, response times will increase, as clients will need to wait for an available slot. This can result in noticeable delays, especially if the queue becomes large.

3. **Potential Timeouts:**
   - If the query stays in the queue for too long and isn't executed within a reasonable timeframe, it may time out. This can lead to errors or failures in the client making the request

### How to Change the `query-max-concurrency` Setting

If you find that the default setting of **4** is either too low or too high for your environment, you can adjust it by modifying the Prometheus configuration file.

1. **Locate the Prometheus configuration file:**
   Typically, this is the `prometheus.yml` file, but the `query-max-concurrency` setting can also be configured in the Prometheus startup flags.

2. **Modify the Setting:**
   In the prometheus-entrypoint.sh file set the `query-max-concurrency` flag to the desired value.
   ```bash
    "$PARENT_DIR/prometheus/prometheus" \
    --config.file="$PARENT_DIR/config/prometheus/prometheus.yml" \
    --storage.tsdb.path="$PARENT_DIR/prometheus_data" \
    --storage.tsdb.allow-overlapping-blocks \
    --query.lookback-delta="15m" \
    --web.console.templates="$PARENT_DIR/config/prometheus/consoles" \
    --web.console.libraries="$PARENT_DIR/config/prometheus/console_libraries" \
    --web.enable-remote-write-receiver \
    --web.enable-admin-api \
    --query.max-concurrency=8
   ```

### Why You Might Want to Change the Default Concurrency Setting

While the default value of **4** is a reasonable starting point, there are several scenarios where you may want to change it:

1. **Higher Traffic:**
   - If you have a high volume of simultaneous queries, you may want to increase the concurrency limit

2. **Resource Availability:**
   - If your system has more CPU cores and memory available, you can safely increase the concurrency limit to fully utilize your hardware.

3. **I/O Bound Queries:**
   - If your queries are I/O-bound, you might want to increase concurrency slightly to **take advantage of I/O waiting times**
