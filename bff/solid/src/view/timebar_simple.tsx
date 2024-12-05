import { flip } from "@floating-ui/dom";
import * as chrono from "chrono-node";
import {
  For,
  type JSX,
  Match,
  Show,
  Switch,
  batch,
  createEffect,
  createSignal,
  on,
  onCleanup,
  untrack,
} from "solid-js";
import { Popover } from "solid-simple-popover";
import { queryEndpointDataSimple } from "~/http_simple";
import { ApiEndpoint } from "~/state";
import { contextState } from "../context_state";
import { cssSelectorGeneral } from "./cube_activity";

interface ITimebarSimpleProps {
  class?: string;
}

export function TimebarSimple(props: ITimebarSimpleProps) {
  const { state } = contextState();
  const [error, setError] = createSignal<string | null>(null);
  const [dirty, setDirty] = createSignal<boolean>(false);

  return (
    <section
      class={`mb-3 flex flex-col lg:flex-row items-center gap-4 ${props.class}`}
    >
      <div class="flex flex-row items-center gap-4">
        <div class="flex flex-col xs:flex-row flex-wrap items-center gap-4">
          <Show when={!state.isLive}>
            <TimeSelector
              error={error}
              setError={setError}
              setDirty={setDirty}
            />
          </Show>
          <TimespanSelector />
          <TimespanString error={error} />
          <Show when={!state.isLive}>
            <UpdateButton />
          </Show>
          <LiveIndicator setError={setError} />
          {/*
          <IntervalSelector />
          */}
        </div>
      </div>
    </section>
  );

  function UpdateButton() {
    const { state } = contextState();
    createEffect(
      on([() => state.chronoRaw], () => {
        if (!state.chronoRaw.length) return;
        setDirty(true);
      }),
    );
    return (
      <button
        type="button"
        disabled={!state.chronoInterpreted}
        onClick={() => {
          queryUpdate();
          setDirty(false);
        }}
        class="text-sm px-2.5 py-2 rounded-lg hover:bg-zinc-300 dark:hover:bg-zinc-700 disabled:opacity-50 disabled:cursor-not-allowed"
        classList={{
          "bg-yellow-400 text-black font-semibold":
            !!state.chronoInterpreted && dirty(),
          "bg-zinc-200 dark:bg-zinc-800": !(dirty() && state.chronoInterpreted),
        }}
      >
        Update
      </button>
    );
  }
}

interface LiveProps {
  setError: (error: string | null) => void;
}

function LiveIndicator(props: LiveProps) {
  const { state, setState } = contextState();
  const [intervalId, setIntervalId] = createSignal<number | null>(null);

  function runLiveUpdate() {
    batch(() => {
      setState("chronoRaw", `${getTimespanLabel(state.timespan_ms)} ago`);
      setState("chronoInterpreted", chrono.parseDate(state.chronoRaw));
      queryUpdate();
    });
  }

  createEffect(
    on(
      [
        () => state.interval_ms,
        () => state.timespan_ms,
        () => state.isLive, //
      ],
      () => {
        if (intervalId()) {
          window.clearInterval(intervalId() || undefined);
          setIntervalId(null);
        }
        if (state.isLive) {
          const id = window.setInterval(runLiveUpdate, state.interval_ms);
          setIntervalId(id);
          runLiveUpdate();
        }
      },
    ),
  );

  onCleanup(() => {
    if (intervalId()) {
      window.clearInterval(intervalId() || undefined);
      setIntervalId(null);
    }
  });

  return (
    <label
      class="flex items-center gap-0.5 border border-yellow-300 dark:border-0 dark:border-green-500 p-2.5 rounded-md bg-yellow-200 text-black font-semibold leading-none"
      classList={{
        "bg-yellow-200": state.isLive,
        "bg-zinc-500": !state.isLive,
      }}
    >
      <input
        type="checkbox"
        checked={state.isLive}
        onChange={(e) => {
          batch(() => {
            if (e.target.checked) {
              props.setError(null);
              setState(
                "chronoRaw",
                `${getTimespanLabel(state.timespan_ms)} ago`,
              );
              setState("chronoInterpreted", chrono.parseDate(state.chronoRaw));
            }
            setState("isLive", e.target.checked);
          });
        }}
      />
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
    </label>
  );
}

function formatTimeRange(start: Date, end: Date): string {
  const startDate = start.toLocaleDateString();
  const endDate = end.toLocaleDateString();
  const startTime = start.toLocaleTimeString();
  const endTime = end.toLocaleTimeString();

  if (startDate === endDate) {
    return `${startDate} ${startTime} to ${endTime}`;
  }
  return `${startDate} ${startTime} to ${endDate} ${endTime}`;
}

interface TimeSelectorProps {
  error: () => string | null;
  setError: (error: string | null) => void;
  setDirty: (arg: boolean) => void;
}

function TimeSelector(props: TimeSelectorProps) {
  const { state, setState } = contextState();

  const examples = [
    // "now",
    "last tuesday 3:15pm",
    // "yesterday at 2pm",
    "2h ago",
    // "3 days ago",
  ];

  createEffect(() => {
    const input = state.chronoRaw;
    if (!input) {
      setState("chronoInterpreted", null);
      props.setError(null);
      return;
    }

    const parsed = chrono.parseDate(input);
    if (!parsed) {
      setState("chronoInterpreted", null);
      props.setError("Unable to parse date/time");
      return;
    }

    const now = new Date();
    const fourteenDaysAgo = new Date(now.getTime() - 14 * 24 * 60 * 60 * 1000);

    if (parsed > now) {
      setState("chronoInterpreted", null);
      props.setError("Cannot set future dates");
      return;
    }

    if (parsed < fourteenDaysAgo) {
      setState("chronoInterpreted", null);
      props.setError("Cannot set dates more than 14 days in the past");
      return;
    }

    setState("chronoInterpreted", parsed);
    props.setError(null);
  });

  return (
    <>
      <form
        class="flex flex-row items-center gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          queryUpdate();
          props.setDirty(false);
        }}
      >
        <div class="flex flex-col gap-2">
          <input
            type="text"
            value={state.chronoRaw}
            onInput={(e) => setState("chronoRaw", e.currentTarget.value)}
            placeholder="Enter time (e.g. '2 hours ago')"
            class="text-sm px-2.5 py-2 rounded-lg bg-zinc-200 dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700"
          />
          <Show when={!state.chronoRaw || props.error()}>
            <div class="text-xs text-zinc-500">
              Examples: {examples.join(", ")}
            </div>
          </Show>
        </div>
      </form>
    </>
  );
}

interface TimespanStringProps {
  error: () => string | null;
}

function TimespanString(props: TimespanStringProps) {
  const { state } = contextState();
  return (
    <div class="text-sm">
      <Switch fallback={<div class="text-red-500">{props.error()}</div>}>
        <Match when={!props.error() && state.chronoInterpreted}>
          {(chronoInterpreted) => (
            <div class="text-green-600 dark:text-green-400">
              <p>
                Time range:{" "}
                <span
                  class={`text-2xs text-neutral-700 dark:text-neutral-300 ${
                    Object.getOwnPropertyNames(
                      state.apiThrottle.requestInFlight,
                    ).length
                      ? "visible"
                      : "invisible"
                  }`}
                >
                  Updating
                </span>
              </p>
              <p>
                {formatTimeRange(
                  chronoInterpreted(),
                  new Date(chronoInterpreted().getTime() + state.timespan_ms),
                )}
              </p>
              <Show
                when={
                  (state.apiThrottle.needDataFor === ApiEndpoint.activity &&
                    !state.activityCube.cubeData.length) ||
                  (state.apiThrottle.needDataFor === ApiEndpoint.metric &&
                    !state.metricData.length) ||
                  (state.apiThrottle.needDataFor ===
                    ApiEndpoint.prometheus_metrics &&
                    !state.prometheusMetricsData.length)
                }
              >
                <span class="text-yellow-600 dark:text-yellow-400">
                  No data available for this time range
                </span>
              </Show>
            </div>
          )}
        </Match>
      </Switch>
    </div>
  );
}

function TimespanSelector() {
  const { setState } = contextState();
  const id = "timespanSelector";

  return (
    <>
      <ViewSelector
        name="Timespan"
        property="timespan_ms"
        id={id}
        options={optionsTimespan}
        onClick={(record) => () =>
          batch(() => {
            setState("timespan_ms", record.ms);
            untrack(() => setState("interval_ms", record.ms2));
            queryUpdate();
          })
        }
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
  property: "interval_ms" | "timespan_ms";
  onClick: (
    arg0: RecordClickHandler,
  ) => JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
  options: RecordClickHandler[];
  id: string;
  class?: string;
}

function ViewSelector(props: PropsViewSelector) {
  const { state } = contextState();
  const id = props.id;

  return (
    <>
      <button
        type="button"
        id={id}
        class={`flex gap-2 text-sm px-2.5 py-2 rounded-lg ${cssSelectorGeneral} ${props.class}`}
      >
        <span class="whitespace-pre me-2">{props.name}:</span>
        <span class="text-fuchsia-500">
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
                type="button"
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
  const id = "intervalSelector_simple";
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
          (record) => state.timespan_ms / record.ms <= 350,
        )}
        onClick={(record) => () =>
          batch(() => {
            setState("interval_ms", record.ms);
            queryUpdate();
          })
        }
        class={props.class}
      />
    </>
  );
}

function queryUpdate() {
  const { state, setState } = contextState();
  if (state.chronoInterpreted) {
    const time_begin = state.chronoInterpreted?.getTime();
    if (!time_begin) return;
    setState("time_begin", time_begin);
    if (state.apiThrottle.needDataFor)
      queryEndpointDataSimple(state.apiThrottle.needDataFor);
  }
}

export const optionsTimespan = [
  { ms: 14 * 24 * 60 * 60 * 1000, label: "14d", ms2: 60 * 60 * 1000 },
  { ms: 7 * 24 * 60 * 60 * 1000, label: "7d", ms2: 30 * 60 * 1000 },
  { ms: 2 * 24 * 60 * 60 * 1000, label: "2d", ms2: 30 * 60 * 1000 },
  { ms: 1 * 24 * 60 * 60 * 1000, label: "1d", ms2: 30 * 60 * 1000 },
  { ms: 12 * 60 * 60 * 1000, label: "12h", ms2: 30 * 60 * 1000 },
  { ms: 6 * 60 * 60 * 1000, label: "6h", ms2: 10 * 60 * 1000 },
  { ms: 3 * 60 * 60 * 1000, label: "3h", ms2: 5 * 60 * 1000 },
  { ms: 1 * 60 * 60 * 1000, label: "1h", ms2: 60 * 1000 },
  { ms: 15 * 60 * 1000, label: "15m", ms2: 10 * 1000 },
  { ms: 2 * 60 * 1000, label: "2m", ms2: 5 * 1000 },
];

function getTimespanLabel(ms: number): string {
  const option = optionsTimespan.find((opt) => opt.ms === ms);
  return option?.label || `${ms / 60000}m`;
}
