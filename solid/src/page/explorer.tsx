import { contextState } from "../context_state";
import { Echarts3 } from "../view/echarts3";
import { ViewTable } from "../view/table";

export function PageExplorer(props: any) {
  const page = "pageExplorer";
  const { state } = contextState();

  return (
    <>
      <section class="grid grid-cols-1 sm:grid-cols-2 gap-4 sm:gap-x-8">
        <div class="flex flex-col overflow-hidden">
          <ViewTable />
        </div>
        <div class="flex flex-col h-64">
          <Echarts3 />
        </div>
      </section>
    </>
  );
}
