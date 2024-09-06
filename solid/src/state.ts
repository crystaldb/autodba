import { batch } from "solid-js";
import { createStore } from "solid-js/store";

export enum DimensionName {
  none = "none",
  query = "query",
  wait_event_name = "wait_event_name",
  time = "time",
  client_addr = "client_addr",
  usename = "usename",
  backend_type = "backend_type",
  application_name = "application_name",
  datname = "datname",
}

export function listDimensionTabNames() {
  return [
    [DimensionName.wait_event_name, "Wait"],
    [DimensionName.query, "Sql"],
    [DimensionName.client_addr, "Host"],
    [DimensionName.usename, "User"],
    [DimensionName.backend_type, "Session type"],
    [DimensionName.application_name, "Application"],
    [DimensionName.datname, "Database"],
  ];
}

export enum DimensionField {
  uiLegend = "uiLegend",
  uiDimension1 = "uiDimension1",
  uiFilter1 = "uiFilter1",
}

export type CubeData = {
  metric: Partial<Record<DimensionName, string>>;
  values: { timestamp: number; value: number }[];
}[];

export type State = {
  cubeActivity: {
    cubeData: CubeData;
    uiLegend: DimensionName;
    uiDimension1: DimensionName;

    uiFilter1: DimensionName;
    uiFilter1Value?: string;
    limit: number;
  };
  database_instance: {
    dbidentifier: string;
    engine: string;
    engine_version: string;
    instance_class: string;
  };
  database_list: string[];
  healthData: {
    cpu: number[];
    time: number[];
  };
  metricData: any[];

  interval_ms: number;
  range_begin: number;
  range_end: number;
  time_begin_ms: number;
  time_end_ms: number;
  window_begin_ms: number;
  window_end_ms: number;
};

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

const appZero = +new Date();

const [state, setState]: [State, any] = createStore({
  cubeActivity: {
    cubeData: [],
    limit: 15,
    uiLegend: DimensionName.wait_event_name,
    uiDimension1: DimensionName.time,
    // uiDimension1: DimensionName.query,
    // uiDimension1: DimensionName.usename,
    uiFilter1: DimensionName.none,
    uiFilter1Value: undefined,
  },
  metricData: [],
  healthData: {
    cpu: [],
    time: [],
  },
  database_instance: {
    dbidentifier: "",
    engine: "",
    engine_version: "",
    instance_class: "",
  },
  database_list: [],
  interval_ms: 50 * 1000, // 5 seconds
  range_begin: 0.0,
  range_end: 100.0,
  time_begin_ms: appZero - 15 * 60 * 1000,
  time_end_ms: appZero,
  window_begin_ms: appZero - 15 * 60 * 1000,
  window_end_ms: appZero,
});

export function useState(): { state: State; setState: any } {
  return { state, setState };
}

export const datazoomEventHandler = (event: any) => {
  console.log("Chart2 Data Zoom", event);
  batch(() => {
    const range_begin: number = event.start || event.batch?.at(0)?.start || 0.0;
    const range_end: number = event.end || event.batch?.at(0)?.end || 100.0;
    setState("range_begin", range_begin);
    setState("range_end", range_end);
    console.log("range", range_begin, range_end);
    const window_begin_ms = Math.floor(
      (state.time_end_ms - state.time_begin_ms) * (range_begin / 100) +
        state.time_begin_ms,
    );
    const window_end_ms = Math.max(
      window_begin_ms,
      Math.ceil(
        (state.time_end_ms - state.time_begin_ms) * (range_end / 100) +
          state.time_begin_ms,
      ),
    );
    console.log("windows", window_begin_ms, window_end_ms);
    setState("window_begin_ms", window_begin_ms);
    setState("window_end_ms", window_end_ms);
  });
};
