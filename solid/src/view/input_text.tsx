import { contextState } from "../context_state";
import { Show } from "solid-js";
import { ViewTitle } from "./common";

interface IViewTextinput {
  title: string;
  type?: string;
}
function ViewTextinput(props: IViewTextinput) {
  const { state, setState } = contextState();

  return (
    <>
      <label class="flex flex-col gap-y-3">
        <Show when={props.title}>
          <ViewTitle title={props.title} />
        </Show>

        <input
          type={props.type || "text"}
          autocomplete="off"
          autocorrect="off"
          spellcheck={false}
          class="font-medium sm:font-base block w-full px-3 py-2.5 sm:py-1.5 sm:text-sm/6 border border-t-zinc-950/15 border-x-zinc-950/20 border-b-zinc-950/30 rounded-md bg-white text-zinc-950 focus:outline-none dark:text-white dark:bg-white/5 dark:border-white/10"
          value={state.str || ""}
          onInput={(e) => {
            const val = e.target.value;
            setState("str", val);
          }}
        />
      </label>
      {state.str}
    </>
  );
}

export { ViewTextinput };
