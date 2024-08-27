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
    [DimensionName.wait_event_name, "Waits"],
    [DimensionName.query, "Sql"],
    [DimensionName.client_addr, "Hosts"],
    [DimensionName.usename, "Users"],
    [DimensionName.backend_type, "Session types"],
    [DimensionName.application_name, "Applications"],
    [DimensionName.datname, "Databases"],
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

    arrActiveSessionCount: number[];
    arrTime: number[];
    arrSql: string[];
    arrWaits: string[];
    arrHosts: string[];
    arrUsers: string[];
    arrSession_types: string[];
    arrApplications: string[];
    arrDatabases: string[];
  };
  data: {
    echart1: number[][];
    echart2a: any[];
    echart2b: any[];
    echart2c: any[];
    echart3: any[];
    cpu: number[];
    time: number[];
  };
  database: {
    name: string;
    engine: string;
    version: string;
    size: string;
    kind: string;
  };
  interval_ms: number;
  timeframe_start_ms: number;
  timeframe_end_ms?: number;
  str: string;
  range_start: number;
  range_end: number;
};

export const datazoom = (
  setState: (arg0: string, arg1: any) => void,
  stateFn: any,
  event: any,
) => {
  console.log("Chart2 Data Zoom", event);
  batch(() => {
    setState("range_start", event.start || event.batch?.at(0)?.start);
    setState("range_end", event.end || event.batch?.at(0)?.end);
  });
};

export type Waits =
  | "LWLock:BufferContent"
  | "LWLock:WALInsert"
  | "Timeout:VaccumDelay"
  | "Timeout:VaccumTruncate"
  | "Client:ClientRead"
  | "IO:WALSync"
  | "Lock:tuple"
  | "LWLock:WALWrite"
  | "Lock:transaactionid"
  | "CPU";

export const listWaits = [
  "LWLock:BufferContent",
  "LWLock:WALInsert",
  "Timeout:VaccumDelay",
  "Timeout:VaccumTruncate",
  "Client:ClientRead",
  "IO:WALSync",
  "Lock:tuple",
  "LWLock:WALWrite",
  "Lock:transaactionid",
  "CPU",
];

export const listWaitsColorsText = [
  "text-blue-300 accent-blue-300",
  "text-green-300 accent-green-300",
  "text-yellow-300 accent-yellow-300",
  "text-red-400 accent-red-400",
  "text-teal-500 accent-teal-500",
  "text-purple-500 accent-purple-500",
  "text-orange-500 accent-orange-500",
  "text-indigo-500 accent-neutral-500",
  "text-fuchsia-500 accent-fuchsia-500",
  "text-green-500 accent-green-500",
];

export const listWaitsColorsBg = [
  "bg-blue-300 accent-blue-300 text-neutral-500",
  "bg-green-300 accent-green-300 text-neutral-500",
  "bg-yellow-300 accent-yellow-300 text-neutral-500",
  "bg-red-400 accent-red-400",
  "bg-teal-500 accent-teal-500",
  "bg-purple-500 accent-purple-500",
  "bg-orange-500 accent-orange-500",
  "bg-indigo-500 accent-neutral-500",
  "bg-fuchsia-500 accent-fuchsia-500",
  "bg-green-500 accent-green-500",
];

const [state, setState] = createStore({
  cubeActivity: {
    cubeData: [],
    limit: 15,
    uiLegend: DimensionName.wait_event_name,
    uiDimension1: DimensionName.query,
    uiFilter1: DimensionName.none,
    uiFilter1Value: undefined,
    arrActiveSessionCount: [],
    arrTime: [],
    arrSql: [],
    arrWaits: [],
    arrHosts: [],
    arrUsers: [],
    arrSession_types: [],
    arrApplications: [],
    arrDatabases: [],
  },
  data: {
    echart1: [[], [], [], [], []],
    echart2a: [],
    echart2b: [],
    echart2c: [],
    echart3: [],
    cpu: [],
    time: [],
  },
  database: {
    name: "",
    engine: "",
    version: "",
    size: "",
    kind: "",
  },
  interval_ms: 30 * 1000, // 30 seconds
  timeframe_start_ms: 0,
  str: "string",
  range_start: 25.0,
  range_end: 100.0,
});

export function useState(): { state: State; setState: any } {
  return { state, setState };
}
