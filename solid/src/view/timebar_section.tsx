import { contextState } from "../context_state";
import { batch, createEffect, For, JSX } from "solid-js";
import { isLiveQueryCube } from "../http";
import { Popover } from "solid-simple-popover";
import { flip } from "@floating-ui/dom";
import { cssThingy } from "./cube_activity";

let debugZero = +new Date();

interface ITimebarSectionProps {
  class?: string;
}

export function TimebarSection(props: ITimebarSectionProps) {
  return (
    <section class={`flex items-center gap-4 ${props.class}`}>
      <TimeframeSelector />

      <IntervalSelector />
      <LiveIndicator />
      {/*
      <TimebarDebugger />
      <EchartsTimebar class="h-12" />
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
      // "1 day",
      // "30m",
      // "30 minutes",
    },
    { ms: 60 * 60 * 1000, label: "last 1h", ms2: 60 * 1000 }, //"1 hour",      "1m", "1 minute"],
    { ms: 15 * 60 * 1000, label: "last 15m", ms2: 10 * 1000 }, //"15 minutes",  "10s", "10 seconds"],
    { ms: 2 * 60 * 1000, label: "last 2m", ms2: 5 * 1000 }, //"2 minutes",  "5s", "5 seconds"],
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

      {/*
      <button
        id={id}
        class={`flex gap-2 text-sm px-2.5 py-2 border-s rounded-lg ${cssThingy}`}
      >
        <span class="whitespace-pre me-2">Timeframe:</span>
        <span class="text-fuchsia-500 w-16">
          {options.find(([ms]) => ms === state.timeframe_ms)?.[1]}
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
          <For each={options}>
            {([ms, label, label2, ms2, label3, label4]) => (
              <button
                class={`flex justify-center gap-2 text-sm px-2.5 py-2 border-s rounded-lg ${cssThingy}`}
                onClick={() =>
                  batch(() => {
                    setState("timeframe_ms", ms);
                    setState("interval_ms", ms2);
                  })
                }
              >
                {label}
              </button>
            )}
          </For>
        </section>
      </Popover>
      */}
    </>
  );
}

interface RecordClickHandler {
  ms: number;
  label: string;
  ms2: number;
}

interface PropsViewSelector {
  name: Element;
  property: "timeframe_ms" | "interval_ms";
  onClick: (
    arg0: RecordClickHandler,
  ) => JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
  options: RecordClickHandler[];
  id: any;
}

function ViewSelector(props: PropsViewSelector) {
  const { state } = contextState();
  const id = props.id;

  return (
    <>
      <button
        id={id}
        class={`flex gap-2 text-sm px-2.5 py-2 border-s rounded-lg ${cssThingy}`}
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
                class={`flex justify-center gap-2 text-sm px-2.5 py-2 border-s rounded-lg ${cssThingy}`}
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

// function IntervalSelector(props: { class?: string }) {
//   const { state, setState } = contextState();
//   // const [openGet, openSet] = createSignal(false);
//
//   // <Show
//   //   when={openGet()}
//   //   fallback={<HiOutlineChevronLeft size="1rem" class="inline ms-1" />}
//   // >
//   //   <HiOutlineChevronDown size="1rem" class="inline ms-1" />
//   // </Show>
//   // <button
//   //   onClick={() => openSet((v) => !v)}
//   //   class={`inline ${props.class}`}
//   // >
//   //   Interval: {state.interval_ms} ms
//   // </button>
//   return (
//     <>
//       <div class={`flex items-center gap-x-3 text-sm ${props.class}`}>
//         <label>Interval</label>
//         <select
//           onChange={(e) => setState({ interval_ms: +e.currentTarget.value })}
//           class="bg-transparent rounded border border-neutral-200 dark:border-neutral-700 text-fuchsia-500 ps-2 pe-8 py-1.5 hover:border-gray-400 focus:outline-none"
//         >
//           <option
//             value="5000"
//             class="appearance-none bg-neutral-100 dark:bg-neutral-800"
//           >
//             Auto
//           </option>
//           <For
//             each={[
//               [1000, "1s"],
//               [5000, "5s"],
//               [10000, "10s"],
//               [20000, "20s"],
//               [30000, "30s"],
//               [60000, "60s"],
//               [300000, "5m"],
//               [900000, "15m"],
//               [1800000, "30m"],
//               [3600000, "60m"],
//             ]}
//           >
//             {([value, label]) => (
//               <option
//                 value={value}
//                 {...(value === state.interval_ms && { selected: true })}
//                 class="appearance-none bg-neutral-100 dark:bg-neutral-800"
//               >
//                 {label}
//               </option>
//             )}
//           </For>
//         </select>
//       </div>
//     </>
//   );
// }

function IntervalSelector() {
  const { setState } = contextState();
  const id = "intervalSelector";
  const options = [
    { ms: 1 * 1000, label: "1s", ms2: 0 },
    { ms: 5 * 1000, label: "5s", ms2: 0 },
    { ms: 10 * 1000, label: "10s", ms2: 0 },
    { ms: 30 * 1000, label: "30s", ms2: 0 },
    { ms: 1 * 60 * 1000, label: "1m", ms2: 0 },
    { ms: 5 * 60 * 1000, label: "5m", ms2: 0 },
    { ms: 15 * 60 * 1000, label: "15m", ms2: 0 },
    { ms: 30 * 60 * 1000, label: "30m", ms2: 0 },
    { ms: 60 * 60 * 1000, label: "1h", ms2: 0 },
  ];

  return (
    <>
      <ViewSelector
        name="Interval"
        property="interval_ms"
        id={id}
        options={options}
        onClick={(record) => (event) =>
          batch(() => {
            setState("interval_ms", record.ms);
          })}
      />

      {/*
      <button
        id={id}
        class={`flex gap-2 text-sm px-2.5 py-2 border-s rounded-lg ${cssThingy}`}
      >
        <span class="whitespace-pre me-2">Timeframe:</span>
        <span class="text-fuchsia-500 w-16">
          {options.find(([ms]) => ms === state.timeframe_ms)?.[1]}
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
          <For each={options}>
            {([ms, label, label2, ms2, label3, label4]) => (
              <button
                class={`flex justify-center gap-2 text-sm px-2.5 py-2 border-s rounded-lg ${cssThingy}`}
                onClick={() =>
                  batch(() => {
                    setState("timeframe_ms", ms);
                    setState("interval_ms", ms2);
                  })
                }
              >
                {label}
              </button>
            )}
          </For>
        </section>
      </Popover>
      */}
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
