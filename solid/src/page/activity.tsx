import { CubeActivity } from "../view/cube_activity";

export function PageActivity() {
  return (
    <section class="flex flex-col gap-y-8">
      <section class="flex flex-col p-4 gap-y-4">
        <Header />
        <CubeActivity />
      </section>
    </section>
  );
}

function Header() {
  return (
    <section class="flex gap-x-2">
      <h2 class="text-xl font-semibold">
        Database Load by Active Session Counts
      </h2>
    </section>
  );
}
