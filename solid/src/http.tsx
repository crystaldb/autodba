import { produce } from "solid-js/store";
import { DimensionName, type State } from "./state";
import { httpFakeDatabase } from "./fake.http";

export async function getDatabaseInfo(
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  // const response = await fetch("/api/query/database", { method: "GET", });

  const json = httpFakeDatabase();

  setState("database", json.database);
  return true;
}

export async function getEndpointData(
  apiEndpoint: string,
  state: State,
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  console.log("getEndpointData");
  if (apiEndpoint === "health") {
    if (!state.database.name) return false;

    const safe_prometheus_11kSampleLimit_ms = 10950 * state.interval_ms;
    const time_start = Math.max(
      state.timeframe_start_ms,
      +new Date() - 15 * 60 * 1000, // query 15 minutes of data
      +new Date() - safe_prometheus_11kSampleLimit_ms, // ensure we do not query too much data.
    );
    const response = await fetch(
      `/api/v1/health?datname=${"(postgres|rdsadmin|template0|template1)" || state.database.name
      }&start=${time_start}&step=${state.interval_ms}ms`,
      {
        method: "GET",
      },
    );

    const json = await response.json();
    if (response.ok) {
      let max_time = 0;
      for (let i = json.length - 1; i >= 0; --i) {
        if (json[i].time_ms > max_time) {
          max_time = json[i].time_ms;
        }
      }
      setState("timeframe_start_ms", Math.max(time_start, max_time));
    }

    const cpu = json.map((item: { cpu: number }) => item.cpu);
    const time = json.map(
      (item: { time_ms: number }) => +new Date(item.time_ms),
    );

    setState(
      "data",
      produce((data: State["data"]) => {
        data.time.push(...time);
        data.cpu.push(...cpu);
      }),
    );
  } else if (apiEndpoint === "activity") {
    if (!state.database.name) return false;
    queryCube(state, setState);
  }
  return true;
}

let queryCubeBusy = false;
export async function queryCube(
  state: State,
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  if (!state.database.name) return false;
  if (queryCubeBusy) return false;

  console.log("queryCube");
  const safe_prometheus_11kSampleLimit_ms = 10950 * state.interval_ms;
  const time_start = Math.max(
    state.cubeActivity.uiDimension1 === DimensionName.time
      ? state.timeframe_start_ms
      : 0,
    +new Date() - 15 * 60 * 1000, // query 15 minutes of data
    +new Date() - safe_prometheus_11kSampleLimit_ms, // ensure we do not query too much data.
  );

  queryCubeBusy = true;
  const response = await fetch(
    `/api/v1/activity?database_list=${"(postgres|rdsadmin|template0|template1)" || state.database.name
    }&start=${time_start}&end=${state.timeframe_end_ms || +new Date()}&step=${state.interval_ms
    }ms&limit=${state.cubeActivity.limit}&legend=${state.cubeActivity.uiLegend
    }&dim=${state.cubeActivity.uiDimension1}&filterdim=${state.cubeActivity.uiFilter1 !== DimensionName.none
      ? state.cubeActivity.uiFilter1
      : ""
    }&filterdimselected=${encodeURIComponent(
      state.cubeActivity.uiFilter1 !== DimensionName.none
        ? state.cubeActivity.uiFilter1Value || ""
        : "",
    )}`,
    {
      method: "GET",
    },
  );
  queryCubeBusy = false;

  const json = await response.json();

  if (response.ok) {
    let values = json[json.length - 1].values;
    let max_time = values[values.length - 1].timestamp;
    console.log("time delta", +new Date() - max_time);
    setState("timeframe_start_ms", Math.max(time_start, max_time));

    setState(
      "cubeActivity",
      produce((cubeActivity: State["cubeActivity"]) => {
        cubeActivity.cubeData = json;
      }),
    );
  }
  return true;
}
