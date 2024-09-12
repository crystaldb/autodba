import { contextState } from "../context_state";
import { batch, createEffect, For, JSX, Show } from "solid-js";
import { isLiveQueryCube } from "../http";
import { Popover } from "solid-simple-popover";
import { flip } from "@floating-ui/dom";
import { cssSelectorGeneral } from "./cube_activity";
import { EchartsTimebar } from "./echarts_timebar";

let debug = true;
let debugZero = +new Date();

interface ITimebarSectionProps {
  class?: string;
}

export function TimebarSection(props: ITimebarSectionProps) {
  const { state } = contextState();
  return (
    <section
      class={`flex flex-col sm:flex-row items-center gap-4 ${props.class}`}
    >
      <LiveIndicator />
      <div class="flex flex-col lg:flex-row items-center gap-4">
        <TimeframeSelector />
        <IntervalSelector class="self-stretch" />
      </div>
      <EchartsTimebar class="h-12 min-w-[calc(16rem)] max-w-[calc(1280px-38rem)] w-[calc(100vw-38rem)] xs:w-[calc(100vw-25rem)]" />
      <Show when={debug}>
        <section class="flex flex-col leading-none text-2xs">
          <p>{JSON.stringify(state.api.needDataFor)}</p>
          <p>{JSON.stringify(state.api.inFlight)}</p>
          <p>
            {state.api.busyWaiting}, {state.api.busyWaitingCount}
          </p>
        </section>
      </Show>
      {/*
      <TimebarDebugger />
      */}
    </section>
  );
}

function TimeframeSelector() {
  const { state, setState } = contextState();
  const id = "timeframeSelector";
  const options = [
    {
      ms: 24 * 60 * 60 * 1000,
      label: "last 1d",
      ms2: 30 * 60 * 1000,
    },
    { ms: 12 * 60 * 60 * 1000, label: "last 12h", ms2: 30 * 60 * 1000 },
    { ms: 6 * 60 * 60 * 1000, label: "last 6h", ms2: 10 * 60 * 1000 },
    { ms: 3 * 60 * 60 * 1000, label: "last 3h", ms2: 5 * 60 * 1000 },
    { ms: 1 * 60 * 60 * 1000, label: "last 1h", ms2: 60 * 1000 },
    { ms: 15 * 60 * 1000, label: "last 15m", ms2: 10 * 1000 },
    { ms: 2 * 60 * 1000, label: "last 2m", ms2: 5 * 1000 },
  ];

  createEffect(() => {
    const timeframe_ms = state.timeframe_ms;
    batch(() => {
      // console.log("update time");
      setState("time_begin_ms", () => state.time_end_ms - timeframe_ms);
      setState("window_begin_ms", () => state.time_end_ms - timeframe_ms);
    });
  });

  return (
    <>
      <ViewSelector
        name="Timeframe"
        property="timeframe_ms"
        id={id}
        options={options}
        onClick={(record) => () =>
          batch(() => {
            setState("timeframe_ms", record.ms);
            setState("interval_ms", record.ms2);
          })}
      />
    </>
  );
}

interface RecordClickHandler {
  ms: number;
  label: string;
  ms2: number;
}

interface PropsViewSelector {
  name: string;
  property: "timeframe_ms" | "interval_ms";
  onClick: (
    arg0: RecordClickHandler,
  ) => JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
  options: RecordClickHandler[];
  id: any;
  class?: string;
}

function ViewSelector(props: PropsViewSelector) {
  const { state } = contextState();
  const id = props.id;

  return (
    <>
      <button
        id={id}
        class={`flex gap-2 text-sm px-2.5 py-2 rounded-lg ${cssSelectorGeneral} ${props.class}`}
      >
        <span class="whitespace-pre me-2">{props.name}:</span>
        <span class="text-fuchsia-500 w-16">
          {props.options.find(({ ms }) => ms === state[props.property])?.label}
        </span>
      </button>
      <Popover
        triggerElement={`#${id}`}
        autoUpdate
        computePositionOptions={{
          placement: "bottom-end",
          middleware: [flip()],
        }}
        // sameWidth
        dataAttributeName="data-open"
      >
        <section class="floating width60 grid grid-cols-1">
          <For each={props.options}>
            {(record) => (
              <button
                class={`flex justify-center gap-2 text-sm px-2.5 py-2 rounded-lg ${cssSelectorGeneral}`}
                classList={{
                  "text-fuchsia-500": state[props.property] === record.ms,
                }}
                onClick={props.onClick(record)}
              >
                {record.label}
              </button>
            )}
          </For>
        </section>
      </Popover>
    </>
  );
}

interface PropsIntervalSelector {
  class?: string;
}

function IntervalSelector(props: PropsIntervalSelector) {
  const { state, setState } = contextState();
  const id = "intervalSelector";
  const options = [
    { ms: 1 * 1000, label: "1s", ms2: 0 },
    { ms: 5 * 1000, label: "5s", ms2: 0 },
    { ms: 10 * 1000, label: "10s", ms2: 0 },
    { ms: 30 * 1000, label: "30s", ms2: 0 },
    { ms: 1 * 60 * 1000, label: "1m", ms2: 0 },
    { ms: 5 * 60 * 1000, label: "5m", ms2: 0 },
    { ms: 10 * 60 * 1000, label: "10m", ms2: 0 },
    { ms: 15 * 60 * 1000, label: "15m", ms2: 0 },
    { ms: 30 * 60 * 1000, label: "30m", ms2: 0 },
    { ms: 1 * 60 * 60 * 1000, label: "1h", ms2: 0 },
  ];

  return (
    <>
      <ViewSelector
        name="Interval"
        property="interval_ms"
        id={id}
        options={options.filter(
          (record) => state.timeframe_ms / record.ms <= 350,
        )}
        onClick={(record) => () =>
          batch(() => {
            setState("interval_ms", record.ms);
          })}
        class={props.class}
      />
    </>
  );
}

function LiveIndicator() {
  const { state } = contextState();
  return (
    <div
      class="border border-yellow-300 dark:border-0 dark:border-green-500 px-2.5 py-2.5 rounded-md bg-yellow-200 text-black font-semibold leading-none"
      classList={{ invisible: !isLiveQueryCube(state) }}
    >
      LIVE
    </div>
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
