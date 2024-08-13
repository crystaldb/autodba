import { contextState } from "../context_state";
import { Echarts2 } from "../view/echarts2";
import { Echarts3 } from "../view/echarts3";

export function PageMetric(props: any) {
  const page = "pageMetric";
  const { state } = contextState();

  return (
    <>
      <section class="flex flex-col gap-y-12">
        <div class="flex flex-col p-4 gap-y-4 rounded-xl bg-neutral-100 dark:bg-neutral-800">
          <h2 class="font-medium">Database metrics</h2>
          <section class="grid grid-cols-1 xs:grid-cols-2 md:grid-cols-3 gap-4">
            <Echarts2
              title="Sessions (sessions)"
              dataA={state.data.echart2a}
              dataB={state.data.echart2b}
              dataC={state.data.echart2c}
              class="h-80 p-4 rounded-lg bg-neutral-100 dark:bg-neutral-950"
            />
            <Echarts2
              title="Connections utilization (%)"
              dataA={state.data.echart2a}
              dataB={state.data.echart2b}
              dataC={state.data.echart2c}
              class="h-80 p-4 rounded-lg bg-neutral-100 dark:bg-neutral-950"
            />
            <Echarts2
              title="Transactions in progress (per second)"
              dataA={state.data.echart2a}
              dataB={state.data.echart2b}
              dataC={state.data.echart2c}
              class="h-80 p-4 rounded-lg bg-neutral-100 dark:bg-neutral-950"
            />
          </section>
        </div>
        <div class="flex flex-col p-4 gap-y-4 rounded-xl bg-neutral-100 dark:bg-neutral-800">
          <h2 class="font-medium">Database metrics</h2>
          <div class="h-80 p-4 rounded-lg bg-neutral-100 dark:bg-neutral-950">
            <Echarts2
              title="Tuples: DML (tuples per second)"
              dataA={state.data.echart2a}
              dataB={state.data.echart2b}
              dataC={state.data.echart2c}
              class="h-80 p-4 rounded-lg bg-neutral-100 dark:bg-neutral-950"
            />
          </div>
        </div>
      </section>
    </>
  );
}
