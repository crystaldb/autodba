import { createMemo, createResource } from "solid-js";
import { contextState } from "../context_state";
import { CubeActivity } from "../view/cube_activity";
import { queryCubeIfLive } from "../http";

export function PageActivity() {
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
      <h2 class="text-xl font-semibold">Database Active Session Counts</h2>
    </section>
  );
}
