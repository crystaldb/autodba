import { produce } from "solid-js/store";
import {
  allowInFlight,
  ApiEndpoint,
  clearBusyWaiting,
  clearInFlight,
  DimensionName,
  setInFlight,
  type State,
} from "./state";
import { batch } from "solid-js";

let debugFirstTimestamp: number = 0;
let debugZero = +new Date();

export async function queryDatabaseInstanceInfo(
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  const response = await fetch("/api/v1/info", { method: "GET" });

  const json = await response.json();
  if (!response.ok) return false;
  setState("database_instance", json || {});
  return true;
}

export async function queryDatabaseList(
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  const response = await fetch("/api/v1/databases", { method: "GET" });

  const json = await response.json();
  if (!response.ok) return false;

  setState("database_list", json || []);
  return true;
}

export async function queryEndpointData(
  apiEndpoint: ApiEndpoint,
  state: State,
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  if (apiEndpoint === ApiEndpoint.activity) return queryCube(state, setState);
  return queryStandardEndpoint(apiEndpoint, state, setState);
}

export function isLiveQueryCube(state: State): boolean {
  if (!state.database_list.length) return false;
  if (state.range_end !== 100) return false;
  return true;
}

export async function queryEndpointDataIfLive(
  apiEndpoint: ApiEndpoint,
  state: State,
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  if (!isLiveQueryCube(state)) return false;
  if (apiEndpoint === ApiEndpoint.activity) return queryCube(state, setState);
  return queryStandardEndpoint(apiEndpoint, state, setState);
}

export async function queryCube(
  state: State,
  setState: (arg0: string, arg1: any, arg2?: any) => void,
  time_begin?: number,
  time_end?: number,
): Promise<boolean> {
  if (!state.database_list.length) return false;
  if (!allowInFlight(ApiEndpoint.activity)) return false;

  const safe_prometheus_11kSampleLimit_ms = (11000 - 50) * state.interval_ms;
  const dateNow = +new Date();

  const request_time_start =
    // time_begin ||
    Math.max(
      0,
      dateNow - state.timeframe_ms, // query max of 15 minutes of data
      dateNow - safe_prometheus_11kSampleLimit_ms, // ensure we do not query too much data.
      // state.time_begin_ms,
      // state.window_begin_ms,
    );
  const request_time_stop =
    time_end ||
    Math.max(
      0,
      state.window_end_ms,
      state.range_end === 100 ? dateNow : 0,
      // state.time_end_ms,
    ) ||
    dateNow;

  const url = `/api/v1/activity?${
    true
      ? ""
      : `t=${Math.floor(
          (request_time_start - debugZero) / 1000 / 60,
        ).toString()}_${Math.floor(
          (request_time_stop - debugZero) / 1000 / 60,
        ).toString()}_${Math.floor(
          (request_time_start - debugZero) / 1000,
        ).toString()}_${Math.floor(
          (request_time_stop - debugZero) / 1000,
        ).toString()}&`
    //
  }database_list=(${
    state.database_list.join("|") //
  })&start=${
    request_time_start //
  }&end=${
    request_time_stop //
  }&step=${
    state.interval_ms //
  }ms&limitdim=${
    state.cubeActivity.limit //
  }&limitlegend=${
    state.cubeActivity.limit //
  }&legend=${
    state.cubeActivity.uiLegend //
  }&dim=${
    state.cubeActivity.uiDimension1 //
  }&filterdim=${
    state.cubeActivity.uiFilter1 !== DimensionName.none
      ? state.cubeActivity.uiFilter1
      : ""
  }&filterdimselected=${encodeURIComponent(
    state.cubeActivity.uiFilter1 !== DimensionName.none
      ? state.cubeActivity.uiFilter1Value || ""
      : "",
  )}`;
  setInFlight(ApiEndpoint.activity, url);
  const response = await fetch(url, { method: "GET" });
  clearInFlight(ApiEndpoint.activity);

  if (!response.ok) {
    clearBusyWaiting();
    return true;
  }
  const json = await response.json();
  if (!json) {
    clearBusyWaiting();
    return false;
  }

  let timeOldest = request_time_start;
  let timeNewest = request_time_start;
  json.data.forEach((row: { values: { timestamp: number }[] }) => {
    for (let i = 0; i < row.values.length; ++i) {
      const timestamp = row.values[i]?.timestamp;
      if (timestamp < timeOldest) timeOldest = timestamp;
      if (timestamp > timeNewest) timeNewest = timestamp;
    }
  });
  batch(() => {
    // console.log(
    //   "render1: ",
    //   json.data.length,
    //   json.data.reduce((acc: any, row: any) => {
    //     return acc + row.values.length;
    //   }, 0),
    // );
    // // 2000/3000
    // // 3000/4000: crash
    setState(
      "cubeActivity",
      produce((cubeActivity: State["cubeActivity"]) => {
        cubeActivity.cubeData = json.data;
      }),
    );
    if (!state.time_begin_ms)
      setState("time_begin_ms", Math.min(state.time_begin_ms, timeOldest));
    if (!state.window_begin_ms) setState("window_begin_ms", timeOldest);

    setState("time_end_ms", Math.max(timeNewest, dateNow));
    if (!state.window_end_ms || state.range_end == 100)
      setState("window_end_ms", Math.max(timeNewest, dateNow));
    clearBusyWaiting();
  });
  return json;
}

async function queryStandardEndpoint(
  apiEndpoint: string,
  state: State,
  setState: (arg0: string, arg1: any, arg2?: any) => void,
): Promise<boolean> {
  if (apiEndpoint !== ApiEndpoint.metric) return false;
  if (!state.database_list.length) return false;
  if (!allowInFlight(ApiEndpoint.metric)) return false;
  const safe_prometheus_11kSampleLimit_ms = (11000 - 50) * state.interval_ms;
  const request_time_start = Math.max(
    0,
    +new Date() - state.timeframe_ms, // query 15 minutes of data max
    +new Date() - safe_prometheus_11kSampleLimit_ms, // ensure we do not query too much data.
  );
  const url = `/api/v1/${
    apiEndpoint //
  }?datname=(${
    state.database_list.join("|") //
  })&start=${
    request_time_start //
  }&end=${
    +new Date() //
  }&step=${
    state.interval_ms //
  }ms&dbidentifier=${
    state.database_instance.dbidentifier //
  }`;
  setInFlight(ApiEndpoint.metric, url);
  const response = await fetch(url, { method: "GET" });

  clearInFlight(ApiEndpoint.metric);
  if (!response.ok) {
    console.log("Response not ok", response);
    clearBusyWaiting();
    return false;
  }

  const { data, server_now } = await response.json();

  let max_time = 0;
  for (let i = data.length - 1; i >= 0; --i) {
    if (data[i].time_ms > max_time) {
      max_time = data[i].time_ms;
    }
  }
  // ++max_time; // do not query the same data again
  batch(() => {
    if (!debugFirstTimestamp) debugFirstTimestamp = data[0].time_ms;

    // console.log( "json",
    //   (json || []).map((row: { time_ms: any }) => row.time_ms - debugFirstTimestamp),
    //   json,
    // );

    const dataBucketName = apiEndpoint + "Data";
    setState(dataBucketName, () => {
      // let newData = spliceArraysTogetherSkippingDuplicateTimestamps(data, data);
      return data;
    });
    clearBusyWaiting();
  });
  return true;
}

// function spliceArraysTogetherSkippingDuplicateTimestamps(
//   arr1: any[],
//   arr2: any[] = [],
// ): any[] {
//   // in the `arr1` array, starting at the end of the array and looking back up to `arr2.length` items, remove any timestamps from `arr1` that are already present in `arr2`, and then append the new `arr2` array to the end of `arr1`.
//   if (arr1.length === 0) return arr2;
//   if (arr2.length === 0) return arr1;
//   const newTimestamps = new Set(
//     arr2.map((row: { time_ms: any }) => row.time_ms),
//   );
//   const rangeStart = Math.max(0, arr1.length - arr2.length);
//
//   let insertAt = arr1.length;
//   for (let i = arr1.length - 1; i >= rangeStart; --i) {
//     if (newTimestamps.has(arr1[i].time_ms)) {
//       insertAt = i;
//     }
//   }
//   return [...arr1.slice(0, insertAt), ...arr2];
// }
