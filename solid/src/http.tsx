import { produce } from "solid-js/store";
import type { State } from "./state";
import { httpFake, httpFakeDatabase } from "./http.fake";

export async function getDatabaseInfo(
  setState: (arg0: string, arg1: any, arg2?: any) => void
) {
  const response = await fetch("/api/query/database", {
    method: "GET",
  });

  // const jsonIGNORED = await response.json();
  // if (!response.ok || response.ok) { }
  const json = httpFakeDatabase();

  setState("database", json.database);
}

export async function getData(
  apiEndpoint: string,
  setState: (arg0: string, arg1: any, arg2?: any) => void
) {
  const response = await fetch(`/api/query/${apiEndpoint}`, {
    method: "GET",
  });

  // const jsonIGNORED = await response.json();
  // if (!response.ok || response.ok) { }
  const json = httpFake();

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
    })
  );
}
