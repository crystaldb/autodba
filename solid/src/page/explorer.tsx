import { contextState } from "../context_state";
import { EchartsStacked } from "../view/echarts_stacked";
import { Echarts2 } from "../view/echarts2";
import { Echarts3 } from "../view/echarts3";
import { ViewTable } from "../view/table";

export function PageExplorer(props: any) {
  const page = "pageExplorer";
  const { state } = contextState();

  return (
    <>
      <section class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-x-8">
        <div class="flex flex-col p-4 gap-y-4 rounded-xl bg-neutral-100 dark:bg-neutral-800 h-64">
          <EchartsStacked data={state.data.echart1} class="h-64" />
        </div>
        <div class="flex flex-col overflow-hidden">
          <ViewTable />
        </div>
        <div class="flex flex-col h-64">
          <Echarts2
            title="Sessions (sessions)"
            class="h-64"
            dataA={state.data.echart2a}
            dataB={state.data.echart2b}
            dataC={state.data.echart2c}
          />
        </div>
        <div class="flex flex-col h-64">
          <Echarts3 />
        </div>
      </section>
    </>
  );
}
