import { contextState } from "../context_state";
import { Show } from "solid-js";
import { isLiveQueryCube } from "../http";

let debugZero = +new Date();

interface ITimebarSectionProps {
  class?: string;
}

export function TimebarSection(props: ITimebarSectionProps) {
  const { state } = contextState();

  return (
    <section class={`flex items-center ${props.class}`}>
      <Show when={isLiveQueryCube(state)}>
        <div class="border border-yellow-300 dark:border-0 dark:border-green-500 px-2.5 py-2 rounded-md bg-yellow-200 text-black font-semibold leading-none">
          LIVE
        </div>
      </Show>
      <TimebarDebugger />
      {/*
      <EchartsTimebar class="h-12" />
      */}
    </section>
  );
}

function TimebarDebugger() {
  const { state } = contextState();

  return (
    <section class="flex">
      <section class="flex border border-green-500 h-6 ms-[5.5rem] me-[5.2rem]">
        <div
          class="bg-yellow-500 h-full"
          style={{
            width: `${
              ((state.window_begin_ms! - state.time_begin_ms!) /
                (state.time_end_ms! - state.time_begin_ms!)) *
              100
            }%`,
          }}
        ></div>
        <div
          class="bg-green-500 h-full"
          style={{
            width: `${
              ((state.window_end_ms! - state.window_begin_ms!) /
                (state.time_end_ms! - state.time_begin_ms!)) *
              100
            }%`,
          }}
        ></div>
        <div
          class="bg-red-500 h-full"
          style={{
            width: `${
              100 -
              ((state.window_end_ms! - state.time_begin_ms!) /
                (state.time_end_ms! - state.time_begin_ms!)) *
                100
            }%`,
          }}
        ></div>
      </section>
      <aside
        data-testid="DEBUG-timebar"
        class="flex flex-col text-xs text-gray-600 dark:text-gray-400 w-50 shrink-0"
      >
        <p>timeBegin: {state.time_begin_ms - debugZero}</p>
        <p>timeEnd: {state.time_end_ms - debugZero}</p>
        <p>windowBegin: {state.window_begin_ms - debugZero}</p>
        <p>windowEnd: {state.window_end_ms - debugZero}</p>
      </aside>
    </section>
  );
}
