import { For } from "solid-js";
import { contextState } from "../context_state";
import { Echarts2 } from "../view/echarts2";

export function PageMetric(props: any) {
  const page = "pageMetric";
  const { state } = contextState();

  return (
    <>
      <section class="flex flex-col gap-y-12">
        <div class="flex flex-col p-4 gap-y-4 rounded-xl bg-neutral-100 dark:bg-neutral-800">
          <h2 class="font-medium">Database metrics</h2>
          <section class="grid grid-cols-1 xs:grid-cols-2 md:grid-cols-3 gap-4">
            <For
              each={
                [
                  [
                    "Transaction Commits",
                    "",
                    [
                      "transactions_commit",
                      // "transactions_in_progress_active_transactions",
                      // "transactions_in_progress_blocked_transactions",
                      // "transactions_rollback",
                    ],
                  ],
                  [
                    "Transactions",
                    "",
                    [
                      // "transactions_commit",
                      "transactions_in_progress_active_transactions",
                      "transactions_in_progress_blocked_transactions",
                      "transactions_rollback",
                    ],
                  ],
                  ["connection_utilization", "", ["connection_utilization"]],
                  ["cpu_utilization", "", ["cpu_utilization"]],
                  ["deadlocks", "", ["deadlocks"]],
                  [
                    "IOPS",
                    "",
                    [
                      "ebs_current_provisioned_iops",
                      "ebs_read_iops",
                      "ebs_write_iops",
                    ],
                  ],
                  [
                    "Memory",
                    "",
                    [
                      "free_memory",
                      "memory_usage_other_freeable_memory",
                      "memory_usage_shared_memory",
                      "memory_usage_unused_instance_memory",
                    ],
                  ],
                  ["free_storage_space", "", ["free_storage_space"]],
                  [
                    "io_read_throughput",
                    "",
                    [
                      "io_read_throughput",
                      "io_vs_disk_blocks_hit",
                      "io_vs_disk_blocks_read",
                      "io_write_throughput",
                    ],
                  ],
                  [
                    "max_time_idle_in_transaction",
                    "",
                    ["max_time_idle_in_transaction"],
                  ],
                  ["sessions", "", ["sessions"]],
                  [
                    "Tuples",
                    "",
                    [
                      "tuples_dml_deleted",
                      "tuples_dml_inserted",
                      "tuples_dml_updated",
                      "tuples_reads_returned",
                      "tuples_reads_returned_fetched",
                    ],
                  ],
                  [
                    "vacuum_max_used_transaction_ids",
                    "",
                    ["vacuum_max_used_transaction_ids"],
                  ],
                ] as [string, string, string[]][]
              }
            >
              {([title, unit, metricList]: [string, string, string[]]) => (
                <section class="p-4 rounded-lg bg-neutral-100 dark:bg-neutral-950">
                  <h2 class="break-words">{title}</h2>
                  <Echarts2
                    title={title}
                    metricList={metricList}
                    data={state.metricData}
                    class="h-80"
                  />
                </section>
              )}
            </For>
          </section>
        </div>
      </section>
    </>
  );
}
