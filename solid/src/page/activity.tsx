import { createMemo, createResource } from "solid-js";
import { contextState } from "../context_state";
import { CubeActivity } from "../view/cube_activity";
import { queryCube } from "../http";

export function PageActivity() {
  const { state, setState } = contextState();
  const changed = createMemo((changeCount: number) => {
    state.database.name;
    state.cubeActivity.uiLegend;
    state.cubeActivity.uiDimension1;
    state.cubeActivity.uiFilter1;
    state.cubeActivity.uiFilter1Value;
    // state.cubeActivity.uiLegendUnchecked;
    // state.cubeActivity.uiDimension1Unchecked;
    // state.cubeActivity.uiFilter1Unchecked;
    console.log("changed", changeCount);
    return changeCount + 1;
  }, 0);

  createResource(changed, () => {
    queryCube(state, setState);
  });

  return (
    <section class="flex flex-col gap-y-8">
      <section class="flex flex-col p-4 gap-y-4 rounded-xl bg-neutral-100 dark:bg-neutral-800">
        <Header />
        <CubeActivity />
      </section>
    </section>
  );
}

function Header() {
  return (
    <section class="flex gap-x-2">
      <h2 class="text-xl font-semibold">Database load</h2>
    </section>
  );
}
