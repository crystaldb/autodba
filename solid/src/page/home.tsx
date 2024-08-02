import { A } from "@solidjs/router";
import { DarkmodeSelector } from "../view/darkmode";
import { contextState } from "../context_state";
import { children, mergeProps, Show } from "solid-js";
import { Echarts1 } from "../view/echarts1";
import { Echarts2 } from "../view/echarts2";
import { Echarts3 } from "../view/echarts3";
import { ViewTable } from "../view/table";

export function PageHome(props: any) {
  const page = "pageHome";
  const { state } = contextState();

  return (
    <>
      <section class="flex flex-col gap-y-12">
        {/*<section data-testid={page} class="gap-y-12 flex flex-col">
          [Crystal Observability homepage stuff here]
        </section>
        */}
        <div class="flex flex-col h-64">
          <Echarts1 data={state.data.echart1} />
        </div>
        <div class="flex flex-col overflow-hidden">
          <ViewTable />
        </div>
        <div class="flex flex-col h-64">
          <Echarts2
            dataA={state.data.echart2a}
            dataB={state.data.echart2b}
            dataC={state.data.echart2c}
          />
        </div>
        <div class="flex flex-col h-64">
          <Echarts3 />
        </div>
      </section>
      <DarkmodeSelector class="mt-16 mb-4" />
    </>
  );
}
