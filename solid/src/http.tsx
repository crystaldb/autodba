import { Part, produce } from "solid-js/store";
import {
  allowInFlight,
  ApiEndpoint,
  clearBusyWaiting,
  clearInFlight,
  DimensionName,
  getTimeAtPercentage,
  setInFlight,
  type State,
} from "./state";
import { batch } from "solid-js";
import { contextState } from "./context_state";

const magicPrometheusMaxSamplesLimit = 11000;

export async function queryDatabaseInstanceInfo(): Promise<boolean> {
  const { setState } = contextState();
  const response = await fetch("/api/v1/info", { method: "GET" });

  const json = await response.json();
  if (!response.ok) return false;
  setState("database_instance", json || {});
  return true;
}

export async function queryDatabaseList(): Promise<boolean> {
  const { setState } = contextState();
  const response = await fetch("/api/v1/databases", { method: "GET" });

  const json = await response.json();
  if (!response.ok) return false;

  setState("database_list", json || []);
  return true;
}

export function isLive(): boolean {
  const { state } = contextState();
  if (!state.database_list.length) return false;
  if (state.range_end !== 100) return false;
  return true;
}

export async function queryEndpointDataIfLive(
  apiEndpoint: ApiEndpoint,
): Promise<boolean> {
  if (!isLive()) return false;
  return queryEndpointData(apiEndpoint);
}

export async function queryEndpointData(
  apiEndpoint: ApiEndpoint,
): Promise<boolean> {
  return apiEndpoint === ApiEndpoint.activity
    ? queryActivityCube()
    : queryStandardEndpointFullTimeframe(apiEndpoint);
}

async function queryActivityCube(): Promise<boolean> {
  const { state } = contextState();
  return state.activityCube.uiDimension1 === DimensionName.time
    ? queryActivityCubeFullTimeframe()
    : queryActivityCubeTimeWindow();
}

async function queryActivityCubeFullTimeframe(): Promise<boolean> {
  const { state, setState } = contextState();
  if (!state.database_list.length) return false;
  if (!allowInFlight(ApiEndpoint.activity)) return false;

  if (
    state.timeframe_ms / state.interval_ms >=
    magicPrometheusMaxSamplesLimit
  ) {
    console.log(
      "Timeframe too large for Prometheus query (11k samples limit)",
      state.interval_ms,
      state.timeframe_ms,
      (state.timeframe_ms / state.interval_ms).toFixed(1),
    );
    clearBusyWaiting();
    return false;
  }

  const url = `/api/v1/activity?why=cube&database_list=(${
    state.database_list.join("|") //
  })&start=${
    `now-${state.timeframe_ms}ms` //
  }&end=${
    "now" //
  }&step=${
    state.interval_ms //
  }ms&limitdim=${
    state.activityCube.limit //
  }&limitlegend=${
    state.activityCube.limit //
  }&legend=${
    state.activityCube.uiLegend //
  }&dim=${
    state.activityCube.uiDimension1 //
  }&filterdim=${
    state.activityCube.uiFilter1 === DimensionName.none ||
    !state.activityCube.uiFilter1Value
      ? ""
      : state.activityCube.uiFilter1
  }&filterdimselected=${encodeURIComponent(
    state.activityCube.uiFilter1 !== DimensionName.none
      ? state.activityCube.uiFilter1Value || ""
      : "",
  )}`;
  setInFlight(ApiEndpoint.activity, url);
  const response = await fetch(url, { method: "GET" });
  clearInFlight(ApiEndpoint.activity);

  if (!response.ok) {
    clearBusyWaiting();
    return true;
  }
  const { data, server_now } = await response.json();
  if (!data) {
    clearBusyWaiting();
    return false;
  }

  batch(() => {
    setState("server_now", server_now);
    // console.log( "render1: ", json.data.length, json.data.reduce((acc: any, row: any) => { return acc + row.values.length; }, 0),); // 2000/3000 // 3000/4000: crash
    setState(
      "activityCube",
      produce((activityCube: State["activityCube"]) => {
        activityCube.cubeData = data;
      }),
    );

    clearBusyWaiting();
  });
  return data;
}

// const debugZero = +new Date();

async function queryActivityCubeTimeWindow(): Promise<boolean> {
  const { state, setState } = contextState();
  if (!state.database_list.length) return false;
  if (!state.server_now) return false;
  if (!allowInFlight(ApiEndpoint.activity)) return false;

  const request_time_begin = Math.floor(
    getTimeAtPercentage(
      { timeframe_ms: state.timeframe_ms, server_now: state.server_now },
      state.range_begin,
    ),
  );
  const request_time_end = Math.max(
    request_time_begin + 1,
    Math.ceil(
      getTimeAtPercentage(
        { timeframe_ms: state.timeframe_ms, server_now: state.server_now },
        state.range_end,
      ),
    ),
  );

  if (
    (request_time_end - request_time_begin) / state.interval_ms >=
    magicPrometheusMaxSamplesLimit
  ) {
    console.log(
      "Time window too large for Prometheus query (11k samples limit)",
      state.interval_ms,
      request_time_end - request_time_begin,
      ((request_time_end - request_time_begin) / state.interval_ms).toFixed(1),
    );
    clearBusyWaiting();
    return false;
  }

  const url = `/api/v1/activity?why=timewindow&${
    ""
    //   `t=${Math.floor(
    //       (request_time_begin - debugZero) / 1000 / 60,
    //     ).toString()}_${Math.floor(
    //       (request_time_end - debugZero) / 1000 / 60,
    //     ).toString()}_${Math.floor(
    //       (request_time_begin - debugZero) / 1000,
    //     ).toString()}_${Math.floor(
    //       (request_time_end - debugZero) / 1000,
    //     ).toString()}&`
    // //
  }database_list=(${
    state.database_list.join("|") //
  })&start=${
    request_time_begin //
  }&end=${
    request_time_end //
  }&step=${
    state.interval_ms //
  }ms&limitdim=${
    state.activityCube.limit //
  }&limitlegend=${
    state.activityCube.limit //
  }&legend=${
    state.activityCube.uiLegend //
  }&dim=${
    state.activityCube.uiDimension1 //
  }&filterdim=${
    state.activityCube.uiFilter1 === DimensionName.none ||
    !state.activityCube.uiFilter1Value
      ? ""
      : state.activityCube.uiFilter1
  }&filterdimselected=${encodeURIComponent(
    state.activityCube.uiFilter1 !== DimensionName.none
      ? state.activityCube.uiFilter1Value || ""
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

  batch(() => {
    // console.log( "render1: ", json.data.length, json.data.reduce((acc: any, row: any) => { return acc + row.values.length; }, 0),); // 2000/3000 // 3000/4000: crash
    setState(
      "activityCube",
      produce((activityCube: State["activityCube"]) => {
        activityCube.cubeData = json.data;
      }),
    );

    clearBusyWaiting();
  });
  return json;
}

export async function queryFilterOptions(): Promise<boolean> {
  const { state, setState } = contextState();
  if (!state.database_list.length) return false;
  if (!state.server_now) return false;

  const url = `/api/v1/activity?why=filteroptions&database_list=(${
    state.database_list.join("|") //
  })&start=${
    `now-${state.timeframe_ms}ms` //
  }&end=${
    "now" //
  }&step=${
    state.interval_ms //
  }ms&limitdim=${
    state.activityCube.limit //
  }&limitlegend=${
    state.activityCube.limit //
  }&legend=${
    state.activityCube.uiFilter1 //
  }&dim=${
    state.activityCube.uiFilter1 //
  }&filterdim=${
    "" //
  }&filterdimselected=${encodeURIComponent(
    "", //
  )}`;
  const response = await fetch(url, { method: "GET" });

  if (!response.ok) {
    return true;
  }
  const json = await response.json();
  if (!json) {
    return false;
  }

  batch(() => {
    setState(
      "activityCube",
      produce((activityCube: State["activityCube"]) => {
        activityCube.filter1Options = json.data;
      }),
    );
  });
  return json;
}

async function queryStandardEndpointFullTimeframe(
  apiEndpoint: string,
): Promise<boolean> {
  const { state, setState } = contextState();
  if (apiEndpoint !== ApiEndpoint.metric) return false;
  if (!state.database_list.length) return false;
  if (!allowInFlight(ApiEndpoint.metric)) return false;

  if (
    state.timeframe_ms / state.interval_ms >=
    magicPrometheusMaxSamplesLimit
  ) {
    console.log(
      "Timeframe too large for Prometheus query (11k samples limit)",
      state.interval_ms,
      state.timeframe_ms,
      (state.timeframe_ms / state.interval_ms).toFixed(1),
    );
    clearBusyWaiting();
    return false;
  }

  const url = `/api/v1/${
    apiEndpoint //
  }?datname=(${
    state.database_list.join("|") //
  })&start=${
    `now-${state.timeframe_ms}ms` //
  }&end=${
    "now" //
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

  batch(() => {
    setState("server_now", server_now);
    const dataBucketName = (apiEndpoint + "Data") as Part<State, keyof State>;
    setState(dataBucketName, data);

    clearBusyWaiting();
  });
  return true;
}
