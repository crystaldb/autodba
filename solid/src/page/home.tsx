import { A } from "@solidjs/router";
import { contextState } from "../context_state";
import { children, mergeProps, Show } from "solid-js";
import { Echarts1 } from "../view/echarts1";
import { Echarts2 } from "../view/echarts2";
import { Echarts3 } from "../view/echarts3";

export function PageHome(props: any) {
  const page = "pageHome";
  // const { state, setState } = contextState();

  return (
    <>
      <section class="flex flex-col gap-y-12">
        <section data-testid={page} class="gap-y-12 flex flex-col">
          [Crystal Observability homepage stuff here]
        </section>
        <div class="flex flex-col bg-white h-64">
          <Echarts1 />
        </div>
        <div class="flex flex-col bg-white h-64">
          <Echarts2 />
        </div>
        <div class="flex flex-col bg-white h-64">
          <Echarts3 />
        </div>
      </section>
    </>
  );
}
