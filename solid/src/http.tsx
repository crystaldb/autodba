import { produce } from "solid-js/store";
import type { State } from "./state";
import { httpFake, httpFakeCubeActivity, httpFakeDatabase } from "./http.fake";

export interface ArraysCubeActivity {
  arrActiveSessionCount: number[];
  arrTime: number[];
  arrSql: string[];
  arrWaits: string[];
  arrHosts: string[];
  arrUsers: string[];
  arrSession_types: string[];
  arrApplications: string[];
  arrDatabases: string[];
}

export async function getDatabaseInfo(
  setState: (arg0: string, arg1: any, arg2?: any) => void,
) {
  const response = await fetch("/api/query/database", {
    method: "GET",
  });

  // const jsonIGNORED = await response.json();
  // if (!response.ok || response.ok) { }
  const json = httpFakeDatabase();

  setState("database", json.database);
  return true;
}

export async function getEndpointData(
  apiEndpoint: string,
  state: State,
  setState: (arg0: string, arg1: any, arg2?: any) => void,
) {
  if (apiEndpoint === "health") {
    if (!state.database.name) return;

    const step_s = Math.floor(state.interval_ms / 1000);
    const safe_prometheus_11kSampleLimit_ms = 10950 * state.interval_ms;
    const time_start = Math.max(
      state.interval_request_ms,
      +new Date() - safe_prometheus_11kSampleLimit_ms,
    );
    const response = await fetch(
      `/api/v1/health?database_id=${state.database.name}&start=${time_start}&step=${step_s}s`,
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
      setState("interval_request_ms", Math.max(time_start, max_time));
      // console.log(
      //   "Prometheus json : ",
      //   state.interval_request_ms,
      //   json.length,
      //   json
      // );
    }

    const cpu = json.map((item: { cpu: number }) => item.cpu);
    const time = json.map(
      (item: { time_ms: number }) => +new Date(item.time_ms),
    );

    const jsonFake = httpFake();

    setState(
      "data",
      produce((data: State["data"]) => {
        data.time.push(...time);
        data.cpu.push(...cpu);

        data.echart2a.push(...jsonFake.echart2a);
        data.echart2b.push(...jsonFake.echart2b);
        data.echart2c.push(...jsonFake.echart2c);
        data.echart1[0].pop();
        data.echart1[1].pop();
        data.echart1[2].pop();
        data.echart1[3].pop();
        data.echart1[4].pop();
        data.echart1[0].push(...jsonFake.echart1[0]);
        data.echart1[1].push(...jsonFake.echart1[1]);
        data.echart1[2].push(...jsonFake.echart1[2]);
        data.echart1[3].push(...jsonFake.echart1[4]);
        data.echart1[4].push(...jsonFake.echart1[0]);
      }),
    );
  } else if (apiEndpoint === "activity") {
    if (!state.database.name) return;

    const step_s = Math.floor(state.interval_ms / 1000);
    const safe_prometheus_11kSampleLimit_ms = 10950 * state.interval_ms;
    const time_start = Math.max(
      state.interval_request_ms,
      +new Date() - safe_prometheus_11kSampleLimit_ms,
    );
    const response = await fetch(
      `/api/v1/activity?database_id=${state.database.name}&start=${
        state.interval_request_ms
      }&step=${Math.floor(state.interval_ms / 1000)}s&legend=${
        state.cubeActivity.uiDimension1
      }&dim=${state.cubeActivity.uiDimension2}&filterdim=${
        state.cubeActivity.uiDimension3
      }&filterdimselected=${encodeURIComponent(state.cubeActivity.uiFilter3)}`,
      {
        method: "GET",
      },
    );

    // const json = await response.json();

    const json = httpFakeCubeActivity();
    setState(
      "cubeActivity",
      produce((cubeActivity: ArraysCubeActivity) => {
        cubeActivity.arrActiveSessionCount.push(...json.arrActiveSessionCount);
        cubeActivity.arrTime.push(...json.arrTime);
        cubeActivity.arrSql.push(...json.arrSql);
        cubeActivity.arrWaits.push(...json.arrWaits);
        cubeActivity.arrHosts.push(...json.arrHosts);
        cubeActivity.arrUsers.push(...json.arrUsers);
        cubeActivity.arrSession_types.push(...json.arrSession_types);
        cubeActivity.arrApplications.push(...json.arrApplications);
        cubeActivity.arrDatabases.push(...json.arrDatabases);
      }),
    );
  } else {
    let json = httpFake();
    setState(
      "data",
      produce((data: State["data"]) => {
        data.time.push(...json.time);
        data.cpu.push(...json.cpu);
        data.echart2a.push(...json.echart2a);
        data.echart2b.push(...json.echart2b);
        data.echart2c.push(...json.echart2c);
        data.echart1[0].pop();
        data.echart1[1].pop();
        data.echart1[2].pop();
        data.echart1[3].pop();
        data.echart1[4].pop();
        data.echart1[0].push(...json.echart1[0]);
        data.echart1[1].push(...json.echart1[1]);
        data.echart1[2].push(...json.echart1[2]);
        data.echart1[3].push(...json.echart1[4]);
        data.echart1[4].push(...json.echart1[0]);
      }),
    );
  }
}
