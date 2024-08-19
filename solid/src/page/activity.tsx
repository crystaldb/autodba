import { contextState } from "../context_state";
import { CubeActivity } from "../view/cube_activity";
import { EchartsStacked } from "../view/echarts_stacked";
import { ViewTable } from "../view/table";

export function PageActivity(props: any) {
  const page = "pageActivity";
  const { state } = contextState();

  return (
    <section class="flex flex-col gap-y-8">
      <section class="flex flex-col p-4 gap-y-4 rounded-xl bg-neutral-100 dark:bg-neutral-800">
        <Header />
        <CubeActivity class="h-64 min-w-64" />
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
