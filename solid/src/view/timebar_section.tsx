import { contextState } from "../context_state";
import {
  batch,
  createEffect,
  createMemo,
  createSignal,
  For,
  getOwner,
  JSX,
  onCleanup,
  runWithOwner,
  Show,
  untrack,
} from "solid-js";
import { isLive, queryEndpointData, queryEndpointDataIfLive } from "../http";
import { Popover } from "solid-simple-popover";
import { flip } from "@floating-ui/dom";
import { cssSelectorGeneral } from "./cube_activity";
import { EchartsTimebar } from "./echarts_timebar";

let debugThrottling = false;

interface ITimebarSectionProps {
  class?: string;
}

export function TimebarSection(props: ITimebarSectionProps) {
  const { state, setState } = contextState();
  let owner = getOwner();
  let timeout: any;
  let destroyed = false;

  const [eventTimeoutOccurred, setTimeoutOccurred] = createSignal<number>(0);

  const eventForceAnUpdateEvenIfNotLive = createMemo((changeCount: number) => {
    state.force_refresh_by_incrementing;
    state.apiThrottle.needDataFor;
    state.interval_ms;
    state.timeframe_ms;
    state.database_instance.dbidentifier;
    state.activityCube.uiLegend;
    state.activityCube.uiDimension1;
    state.activityCube.uiFilter1;
    state.activityCube.uiFilter1Value;
    console.log("changed_timebar_evenIfNotLive", changeCount);
    return changeCount + 1;
  }, 0);

  // const eventSomethingChangedSoUpdateIfLive = createMemo(
  //   (changeCount: number) => {
  //     // state.cubeActivity.uiFilter1Value;
  //     console.log("changed_timebar", changeCount);
  //     return changeCount + 1;
  //   },
  //   0,
  // );

  const doRestartTheTimeout = () => {
    const interval_ms = state.interval_ms;

    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
    if (destroyed) return;
    timeout = setTimeout(() => {
      runWithOwner(owner, () => {
        setTimeoutOccurred((prev) => prev + 1);
        doRestartTheTimeout();
      });
    }, interval_ms);
  };

  createEffect(() => {
    // eventSomethingChangedSoUpdateIfLive();
    eventForceAnUpdateEvenIfNotLive();
    doRestartTheTimeout();
  });

  let allowOnlyOneInitialQueryToRunAtStartup = true;
  let blockOtherInitialQueryAtStartup = false;

  createEffect((allow) => {
    eventForceAnUpdateEvenIfNotLive();

    if (!allow) return true; // block first run caused by constructor being initialized, but then allow next run

    untrack(() => {
      if (state.apiThrottle.needDataFor) {
        console.log("queryEndpointData_FORCE");
        queryEndpointData(state.apiThrottle.needDataFor, state, setState);
      }
    });
    return true;
  }, allowOnlyOneInitialQueryToRunAtStartup);

  createEffect((allow) => {
    // eventSomethingChangedSoUpdateIfLive();
    eventTimeoutOccurred();

    if (!allow) return true; // block first run caused by constructor being initialized, but then allow next run

    untrack(() => {
      if (state.apiThrottle.needDataFor) {
        console.log("queryEndpointData_IfLive");
        queryEndpointDataIfLive(state.apiThrottle.needDataFor, state, setState);
      }
    });
    return true;
  }, blockOtherInitialQueryAtStartup);

  onCleanup(() => {
    // console.log("CLEANUP interval");
    destroyed = true;
    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
  });

  return (
    <section
      class={`flex flex-col lg:flex-row items-center gap-4 ${props.class}`}
    >
      <div class="flex flex-row items-center gap-4">
        <LiveIndicator />
        <div class="flex flex-col xs:flex-row items-center gap-4">
          <TimeframeSelector />
          <IntervalSelector class="self-stretch" />
        </div>
      </div>
      <EchartsTimebar class="h-12 min-w-[18rem] max-w-[calc(1280px - 39rem)] w-[18rem] xs:w-[24rem] sm:w-[36rem]" />
      <Show when={debugThrottling && state.apiThrottle.requestWaitingCount}>
        <section class="flex flex-col leading-none text-2xs">
          <p>{JSON.stringify(state.apiThrottle.needDataFor)}</p>
          <p>{JSON.stringify(state.apiThrottle.requestInFlight)}</p>
          <p>
            {state.apiThrottle.requestWaiting},{" "}
            {state.apiThrottle.requestWaitingCount}
          </p>
        </section>
      </Show>
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
      classList={{ invisible: !isLive(state) }}
    >
      <span class="invisible">.</span>
      LIVE
      <span
        class={
          Object.getOwnPropertyNames(state.apiThrottle.requestInFlight).length
            ? "visible"
            : "invisible"
        }
      >
        .
      </span>
    </div>
  );
}
