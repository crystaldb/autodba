import { produce } from "solid-js/store";
import { DimensionName, type State } from "./state";
import { httpFakeDatabase } from "./fake.http";

let debugFirstTimestamp: number = 0;

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
  if (apiEndpoint === "activity") {
    if (!state.database.name) return false;
    return await queryCube(state, setState);
  } else if (apiEndpoint === "health" || apiEndpoint === "metric") {
    if (!state.database.name) return false;

    const safe_prometheus_11kSampleLimit_ms = 10950 * state.interval_ms;
    const time_start = Math.max(
      state.timeframe_start_ms,
      +new Date() - 15 * 60 * 1000, // query 15 minutes of data max
      +new Date() - safe_prometheus_11kSampleLimit_ms, // ensure we do not query too much data.
    );
    const response = await fetch(
      `/api/v1/${apiEndpoint}?datname=${
        state.database.datname || state.database.name
      }&start=${time_start}&end=${+new Date()}&step=${
        state.interval_ms
      }ms&dbidentifier=${state.database.name}`,
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
      // ++max_time; // do not query the same data again
      setState("timeframe_start_ms", Math.max(time_start, max_time));
    }

    if (!debugFirstTimestamp) debugFirstTimestamp = json[0].time_ms;

    console.log(
      "json",
      (json || []).map(
        (row: { time_ms: any }) => row.time_ms - debugFirstTimestamp,
      ),
      json,
    );

    const dataBucketName = apiEndpoint + "Data";
    const maxDataPoints = 60 * 15 * 12; // number of 5 second intervals in 15 minutes
    setState(dataBucketName, (data: any[]) => {
      let newData = spliceArraysTogetherSkippingDuplicateTimestamps(data, json);
      if (newData.length > maxDataPoints)
        newData = newData.slice(-maxDataPoints);
      return newData;
    });
    console.log("newdata:", state.metricData.length);
    return true;
  }
  return true;
}

function spliceArraysTogetherSkippingDuplicateTimestamps(
  arr1: any[],
  arr2: any[] = [],
): any[] {
  //immutable version : void //mutable version
  // in the `arr1` array, starting at the end of the array and looking back up to `arr2.length` items, remove any timestamps from `arr1` that are already present in `arr2`, and then append the new `arr2` array to the end of `arr1`.
  // if (arr2.length === 0) return;
  // if (arr1.length === 0) {
  //   arr1.push(...arr2);
  //   return;
  // }
  // immutable version
  if (arr1.length === 0) return arr2;
  if (arr2.length === 0) return arr1;
  const newTimestamps = new Set(
    arr2.map((row: { time_ms: any }) => row.time_ms),
  );
  const rangeStart = Math.max(0, arr1.length - arr2.length);

  let insertAt = arr1.length;
  for (let i = arr1.length - 1; i >= rangeStart; --i) {
    if (newTimestamps.has(arr1[i].time_ms)) {
      insertAt = i;
    }
  }
  // arr1.splice(insertAt, arr1.length - insertAt, ...arr2);
  // immutable version:
  return [...arr1.slice(0, insertAt), ...arr2];
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
    state.cubeActivity.uiDimension1 === DimensionName.time // ? state.timeframe_start_ms
      ? 0
      : 0,
    +new Date() - 15 * 60 * 1000, // query 15 minutes of data
    +new Date() - safe_prometheus_11kSampleLimit_ms, // ensure we do not query too much data.
  );

  queryCubeBusy = true;
  const response = await fetch(
    `/api/v1/activity?database_list=${
      "(postgres|rdsadmin|template0|template1)" || state.database.name
    }&start=${time_start}&end=${state.timeframe_end_ms || +new Date()}&step=${
      state.interval_ms
    }ms&limit=${state.cubeActivity.limit}&legend=${
      state.cubeActivity.uiLegend
    }&dim=${state.cubeActivity.uiDimension1}&filterdim=${
      state.cubeActivity.uiFilter1 !== DimensionName.none
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

  if (response.ok) {
    const json = await response.json();
    if (!json) return false;

    // let values = json[json.length - 1].values;
    // let max_time = values[values.length - 1].timestamp;
    // console.log("time delta", +new Date() - max_time);
    // setState("timeframe_start_ms", Math.max(time_start, max_time));

    setState(
      "cubeActivity",
      produce((cubeActivity: State["cubeActivity"]) => {
        cubeActivity.cubeData = json;
      }),
    );
  }
  return true;
}
