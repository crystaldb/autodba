import { type JSX, batch } from "solid-js";
import { type SetStoreFunction, createStore, produce } from "solid-js/store";
import { queryEndpointData } from "./http";

const dateZero = +new Date();

export type State = {
  apiThrottle: ApiThrottle;
  server_now?: number;

  isLive: boolean;
  time_begin: number; // simple
  timespan_ms: number;
  chronoRaw: string;
  chronoInterpreted: Date | null;
  interval_ms: number;

  timeframe_ms: number;
  range_begin: number;
  range_end: number;
  force_refresh_by_incrementing: number;

  database_list: string[];
  instance_active?: Instance;
  instance_list: Instance[];
  activityCube: ActivityCube;
  // biome-ignore lint/suspicious/noExplicitAny: TODO make this explicit
  metricData: /*eslint-disable */ any[] /*eslint-enable */;
  // biome-ignore lint/suspicious/noExplicitAny: TODO make this explicit
  prometheusMetricsData: /*eslint-disable */ any[] /*eslint-enable */;
};

export type Instance = {
  /** uniquely identifies instance */
  dbIdentifier: string;
  systemId: string;
  systemScope: string;
  systemType: string;
};

export type Anything =
  | number
  | boolean
  | Node
  | JSX.ArrayElement
  | (string & {})
  | null
  | undefined;

export type OrgUser = {
  userId: string;
  name: string;
  email: string;
  roleType: string;
  invitedAt: string;
};

/** ApiType: State used for handling API request throttling
 * CONTEXT: It's possible for some data API requests to take multiple seconds to return. Given that we are polling, we do not want multiple requests to be inflight at the same time, which can overload the backend, and slow down the frontend given how many graphs are rendered. As such, we want to throttle API requests such that we drop/skip any requests that are made while another request is inflight; while also making sure that if 1 or more requests are dropped/skipped because another request is inflight, that the last request (the most recent request) is run as soon as the inflight request completes. This enables the user to miss requests, or to change the page (which requires a new request) without incurring the cost of old, useless requests. To implement this, when each request is made, we check `allowInFlight`, and then either execute the query immediately, or we save the request name as `requestWaiting` and return, effectively dropping/skipping the request for now. If new requests are made, we set `requestWaiting` to the new request name and don't worry about what the previous value was; though, we do increment the `requestWaitingCount` for debugging/observability. As soon as the current inFlight request is completed, whether successfully or with an error, the `requestWaiting` is executed and then set to undefined, along with clearing the `requestWaitingCount`.
 */
export type ApiThrottle = {
  needDataFor?: ApiEndpoint;
  requestInFlight: Record<string, number>;
  requestWaiting?: ApiEndpoint;
  requestWaitingCount?: number;
  requestInFlightUrl?: string;
};

export type ActivityCube = {
  cubeData: CubeData;
  uiLegend: DimensionName;
  uiDimension1: DimensionName;
  filter1Options?: CubeData;

  uiFilter1: DimensionName;
  uiFilter1Value?: string;
  limit: number;
  error?: string;
};

export enum ApiEndpoint {
  activity = "activity",
  metric = "metric",
  prometheus_metrics = "prometheus_metrics",
}

export type DimensionSpec = [DimensionName, string, DimensionRef];
export enum DimensionName {
  none = "",
  query = "query_text",
  wait_event_name = "wait_event_name",
  time = "time",
  client_addr = "client_addr",
  usename = "usename",
  backend_type = "backend_type",
  application_name = "application_name",
  datname = "datname",
}
export enum DimensionRef {
  none = "",
  query = "query_fp",
}

export function listDimensionTabNames(): DimensionSpec[] {
  return [
    [DimensionName.wait_event_name, "Wait", DimensionRef.none],
    [DimensionName.query, "Sql", DimensionRef.query],
    [DimensionName.client_addr, "Host", DimensionRef.none],
    [DimensionName.usename, "User", DimensionRef.none],
    [DimensionName.backend_type, "Session type", DimensionRef.none],
    [DimensionName.application_name, "Application", DimensionRef.none],
    [DimensionName.datname, "Database", DimensionRef.none],
  ];
}

export enum DimensionField {
  uiLegend = "uiLegend",
  uiDimension1 = "uiDimension1",
  uiFilter1 = "uiFilter1",
}
export enum DimensionFieldWithValue {
  uiFilter1Value = "uiFilter1Value",
}
export type DimensionFieldExtended = DimensionField | DimensionFieldWithValue;

export type CubeData = {
  metric: Partial<Record<DimensionName, string>>;
  values: { timestamp: number; value: number }[];
}[];

export type LegendData = {
  metric: Partial<Record<DimensionName, string>>;
  values: { value: number }[];
}[];

export const listColors = [
  {
    text: "text-blue-300 accent-blue-300",
    bg: "bg-blue-300 accent-blue-300 text-neutral-500",
    hex: "#93c5fd", // "bg-blue-300",
  },
  {
    text: "text-green-400 accent-green-400 dark:text-green-300 dark:accent-green-300 font-medium dark:font-normal",
    bg: "bg-green-400 accent-green-400 dark:bg-green-300 dark:accent-green-300 text-neutral-500",
    hex: "#86efac", // "bg-green-300",
  },
  {
    text: "text-yellow-500 accent-yellow-500 dark:text-yellow-300 dark:accent-yellow-300",
    bg: "bg-yellow-500 dark:bg-yellow-300 accent-yellow-500 dark:accent-yellow-300 text-neutral-500",
    hex: "#fde047", // "bg-yellow-300",
  },
  {
    text: "text-red-400 accent-red-400",
    bg: "bg-red-400 accent-red-400",
    hex: "#fca5a5", // "bg-red-400",
  },
  {
    text: "text-teal-500 accent-teal-500",
    bg: "bg-teal-500 accent-teal-500",
    hex: "#5eead4", // "bg-teal-500",
  },
  {
    text: "text-purple-500 accent-purple-500",
    bg: "bg-purple-500 accent-purple-500",
    hex: "#a855f7", // "bg-purple-500",
  },
  {
    text: "text-orange-500 accent-orange-500",
    bg: "bg-orange-500 accent-orange-500",
    hex: "#f97316", // "bg-orange-500",
  },
  {
    text: "text-indigo-500 accent-neutral-500",
    bg: "bg-indigo-500 accent-neutral-500",
    hex: "#6366f1", // "bg-indigo-500",
  },
  {
    text: "text-fuchsia-500 accent-fuchsia-500",
    bg: "bg-fuchsia-500 accent-fuchsia-500",
    hex: "#d946ef", // "bg-fuchsia-500",
  },
  {
    text: "text-green-700 accent-green-700 dark:text-green-500 dark:accent-green-500",
    bg: "bg-green-700 accent-green-700 dark:bg-green-500 dark:accent-green-500",
    hex: "#14b8a6", // "bg-green-500",
  },
  // ... add a bunch of neutral colors for non differentiated colors to avoid eCharts wrap-around
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
  {
    text: "text-gray-500 accent-gray-500",
    bg: "bg-gray-500 accent-gray-500",
    hex: "#6b7280", // "bg-gray-500",
  },
];

const initial_timeframe_ms = 15 * 60 * 1000; // 15 minutes
const initial_interval_ms = 10 * 1000; // 10 seconds

const [state, setState]: [State, SetStoreFunction<State>] = createStore({
  apiThrottle: {
    requestInFlight: {},
  },
  activityCube: {
    cubeData: [],
    limit: 15,
    uiLegend: DimensionName.wait_event_name,
    uiDimension1: DimensionName.time,
    uiFilter1: DimensionName.none,
    uiFilter1Value: undefined,
  },
  metricData: [],
  database_list: [],
  instance_list: [],
  isLive: true,
  time_begin: 0,
  timespan_ms: initial_timeframe_ms,
  timeframe_ms: initial_timeframe_ms,
  interval_ms: initial_interval_ms,
  range_begin: 0.0,
  range_end: import.meta.env.VITE_DEV_MODE === "true" ? 99.5 : 100.0,
  force_refresh_by_incrementing: 0,
  prometheusMetricsData: [],
  chronoRaw: "",
  chronoInterpreted: null,
  // chronoRaw: "15 minutes ago",
  // chronoInterpreted: chrono.parseDate("15 minutes ago"),
} as State);

export function useState(): {
  state: State;
  setState: SetStoreFunction<State>;
} {
  return { state, setState };
}

export const datazoomEventHandler = (
  // biome-ignore lint/suspicious/noExplicitAny: not sure what the event is
  event: /*eslint-disable */ any /*eslint-enable */,
) => {
  console.log("Chart2 Data Zoom", event);
  batch(() => {
    const wasOriginalRangeEndEqualTo100: boolean = state.range_end === 100.0;
    const range_begin: number = event.start || event.batch?.at(0)?.start || 0.0;
    const range_end: number = event.end || event.batch?.at(0)?.end || 100.0;
    setState("range_begin", range_begin);
    setState("range_end", range_end);
    // console.log("range", range_begin, range_end);

    if (!state.apiThrottle.needDataFor) return;

    const rangeBecameLive =
      range_end === 100.0 && !wasOriginalRangeEndEqualTo100;
    const cubeBarsNeedAnUpdate =
      state.activityCube.uiDimension1 !== DimensionName.time &&
      state.apiThrottle.needDataFor === ApiEndpoint.activity;

    if (rangeBecameLive || cubeBarsNeedAnUpdate) {
      // console.log("Forcing a refresh", state.force_refresh_by_incrementing);
      setState("force_refresh_by_incrementing", (prev: number) => prev + 1);
    }
  });
};

export function setBusyWaiting(endpoint: ApiEndpoint) {
  setState(
    "apiThrottle",
    produce((apiThrottle: ApiThrottle) => {
      apiThrottle.requestWaiting = endpoint;
      apiThrottle.requestWaitingCount =
        (apiThrottle.requestWaitingCount || 0) + 1;
    }),
  );
}
export function clearBusyWaiting() {
  const requestWaiting = state.apiThrottle.requestWaiting;
  const requestWaitingCount = state.apiThrottle.requestWaitingCount;
  setState(
    "apiThrottle",
    produce((apiThrottle: ApiThrottle) => {
      // biome-ignore lint/style/noNonNullAssertion: required by SolidJS
      apiThrottle.requestWaiting = undefined!;
      // biome-ignore lint/style/noNonNullAssertion: required by SolidJS
      apiThrottle.requestWaitingCount = undefined!;
    }),
  );
  if (requestWaiting) {
    if (import.meta.env.VITE_DEV_MODE !== "true") {
      console.log(
        "requestWaiting now",
        requestWaitingCount || 0,
        requestWaiting,
      );
      queryEndpointData(requestWaiting);
    }
  }
}

export function allowInFlight(endpoint: ApiEndpoint): boolean {
  if (
    state.apiThrottle.requestInFlight[endpoint] ||
    state.apiThrottle.requestWaiting
  ) {
    setBusyWaiting(endpoint);
    return false;
  }
  return true;
}

export function setInFlight(endpoint: ApiEndpoint, url?: string) {
  setState(
    "apiThrottle",
    produce((apiThrottle: ApiThrottle) => {
      apiThrottle.requestInFlight[endpoint] = +new Date() - dateZero;
      apiThrottle.requestInFlightUrl = url;
    }),
  );
}

export function clearInFlight(endpoint: ApiEndpoint) {
  setState(
    "apiThrottle",
    produce((apiThrottle: ApiThrottle) => {
      // biome-ignore lint/style/noNonNullAssertion: required by SolidJS
      apiThrottle.requestInFlight[endpoint] = undefined!;
    }),
  );
}

export function getTimeAtPercentage(
  state: { timeframe_ms: number; server_now: number },
  numberBetween0And100: number,
): number {
  return (
    state.server_now - state.timeframe_ms * (1 - numberBetween0And100 / 100)
  );
}
