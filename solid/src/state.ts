import { createStore } from "solid-js/store";

export type State = {
  cubeActivity: {
    uiDimension1: string;
    uiDimension2: string;
    uiDimension3: string;
    uiFilter3: string;
    uiCheckedDimension1: string[];
    uiCheckedDimension2: string[];
    uiCheckedDimension3: string[];
    arrActiveSessionCount: number[];
    arrTime: number[];
    arrSql: string[];
    arrWaits: string[];
    arrHosts: string[];
    arrUsers: string[];
    arrSession_types: string[];
    arrApplications: string[];
    arrDatabases: string[];
    // avgSql: string[];
    // avgWaits: string[];
    // avgHosts: string[];
    // avgUsers: string[];
    // avgSession_types: string[];
    // avgApplications: string[];
    // avgDatabases: string[];
    uniqueSql: string[];
    uniqueWaits: string[];
    uniqueHosts: string[];
    uniqueUsers: string[];
    uniqueSession_types: string[];
    uniqueApplications: string[];
    uniqueDatabases: string[];
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
  interval_request_ms: number;
  str: string;
  range_start: number;
  range_end: number;
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

export const listWaitsColors = [
  "text-blue-300 accent-blue-300",
  "text-green-300 accent-green-300",
  "text-yellow-300 accent-yellow-300",
  "text-red-400 accent-red-400",
  "text-teal-500 accent-teal-500",
  "text-purple-500 accent-purple-500",
  "text-orange-500 accent-orange-500",
  "text-neutral-500 accent-neutral-500",
  "text-fuchsia-500 accent-fuchsia-500",
  "text-green-500 accent-green-500",
];

const [state, setState] = createStore({
  cubeActivity: {
    uiDimension1: "waits",
    uiDimension2: "sql",
    // uiDimension2: "time",
    uiDimension3: "sql",
    uiFilter3: "",
    uiCheckedDimension1: listWaits,
    uiCheckedDimension2: [],
    uiCheckedDimension3: [],
    arrActiveSessionCount: [],
    arrTime: [],
    arrSql: [],
    arrWaits: [],
    arrHosts: [],
    arrUsers: [],
    arrSession_types: [],
    arrApplications: [],
    arrDatabases: [],
    // avgSql: [],
    // avgWaits: [],
    // avgHosts: [],
    // avgUsers: [],
    // avgSession_types: [],
    // avgApplications: [],
    // avgDatabases: [],
    uniqueSql: [],
    uniqueWaits: [],
    uniqueHosts: [],
    uniqueUsers: [],
    uniqueSession_types: [],
    uniqueApplications: [],
    uniqueDatabases: [],
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
  interval_ms: 5 * 1000,
  interval_request_ms: 0,
  str: "string",
  range_start: 25.0,
  range_end: 100.0,
});

export function useState(): { state: State; setState: any } {
  return { state, setState };
}
