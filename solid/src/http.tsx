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
  /* const response = await fetch(`/api/query/${apiEndpoint}`, {
   *   method: "GET",
   * }); */

  const response = await fetch(
    `http://localhost:4000/api/v1/health?database_id=mohammad-dashti-rds-1&start=1723740716000&step=5s`,
    {
      method: "GET",
    }
  );

  const json = await response.json();
  if (response.ok) {
    console.log("Prometheus json : ", json);
  }

  // const cpu = json.map((item: { cpu: number }) => Math.floor(item.cpu));
  const cpu = json.map((item: { cpu: number }) => item.cpu);
  const time = json.map((item: { time_ms: number }) => +new Date(item.time_ms));

  const jsonFake = httpFake();

  setState(
    "data",
    produce((data: State["data"]) => {
      data.cpu = cpu;
      data.time = time;
      // data.time.push(...json.time);
      // data.cpu.push(...json.cpu);

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
    })
  );
}
