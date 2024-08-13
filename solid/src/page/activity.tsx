import { contextState } from "../context_state";
import { EchartsStacked } from "../view/echarts_stacked";

export function PageActivity(props: any) {
  const page = "pageActivity";
  const { state } = contextState();

  return (
    <>
      <section class="flex flex-col gap-y-12">
        <div class="flex flex-col p-4 gap-y-4 rounded-xl bg-neutral-100 dark:bg-neutral-800">
          <section class="flex justify-between gap-x-2">
            <h2 class="font-medium">Database load</h2>
            <div class="flex items-center gap-x-3 text-sm">
              <label>Slice by:</label>
              <select class="bg-transparent rounded border border-neutral-200 dark:border-neutral-700 text-fuchsia-500 ps-2 pe-8 py-1.5 hover:border-gray-400 focus:outline-none">
                <option class="appearance-none bg-neutral-100 dark:bg-neutral-800 leading-10">
                  Waits
                </option>
              </select>
            </div>
          </section>
          <EchartsStacked data={state.data.echart1} class="h-64" />
        </div>
      </section>
    </>
  );
}
