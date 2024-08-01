import { produce } from "solid-js/store";
import type { State } from "./state";

let timeB = 0;
let dateTimeDay = "2009/10/18";
let dateTimeHour = 8;

export async function getData(
  setState: (arg0: string, arg1: any, arg2?: any) => void,
) {
  const response = await fetch("/api/data", {
    method: "GET",
  });

  // BEGIN MOCK
  if (!response.ok || response.ok) {
    if (timeB === 0) {
      setState(
        "data",
        produce((data: State["data"]) => {
          data.echart1 = [
            [100, 302, 301, 334, 390, 330, 320],
            [320, 132, 101, 134, 90, 230, 210],
            [220, 182, 191, 234, 290, 330, 310],
            [150, 212, 201, 154, 190, 330, 410],
            [820, 832, 901, 934, 1290, 1330, 1320],
          ];
          data.echart2a = dataA().slice(0, 4);
          data.echart2b = dataB().slice(0, 4);
          data.echart2c = dataC().slice(0, 4);
        }),
      );
    } else {
      setState(
        "data",
        produce((data: State["data"]) => {
          let dataA = data.echart2a;
          let a = dateTimeDay + "\n" + ++dateTimeHour + ":00";
          dataA.push(a);
          let dataB = data.echart2b;
          let b = Math.random() * 10;
          dataB.push(b);
          let dataC = data.echart2c;
          let c = Math.random() * 100;
          dataC.push(c);
          // console.log("a, b, c", a, b, c, dataC.length);

          data.echart1 = [
            [100, 302, 301, 334, 390, 330, tweakValueAt(data.echart1, 0, 6)],
            [320, 132, 101, 134, 90, 230, tweakValueAt(data.echart1, 1, 6)],
            [220, 182, 191, 234, 290, 330, tweakValueAt(data.echart1, 2, 6)],
            [150, 212, 201, 154, 190, 330, tweakValueAt(data.echart1, 3, 6)],
            [820, 832, 901, 934, 1290, 1330, tweakValueAt(data.echart1, 4, 6)],
          ];
          data.echart2a = dataA;
          data.echart2b = dataB;
          data.echart2c = dataC;
        }),
      );
    }
  }
  // END MOCK

  timeB = Date.now();
  return response.json().then((json) => {
    setState("data", json.data);
  });
}

function tweakValueAt(data: number[][], index: number, col: number) {
  const newVal = data[index][col] + Math.floor(Math.random() * 100 + 0.5);
  return Math.max(0, newVal);
}

function dataA() {
  // prettier-ignore
  return [
    "2009/6/12 2:00",
    "2009/6/12 3:00",
  ].map(function (str) {
    return str.replace(" ", "\n");
  });
}

function dataB() {
  // prettier-ignore
  return [
    0.97, 0.96,
  ];
}

function dataC() {
  // prettier-ignore
  return [
    0, 0,
  ];
}
