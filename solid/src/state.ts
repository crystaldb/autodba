import { createStore } from "solid-js/store";

export type State = {
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
  str: string;
  range_start: number;
  range_end: number;
};

const [state, setState] = createStore({
  data: {
    echart1: [[], [], [], [], []],
    echart2a: [],
    echart2b: [],
    echart2c: [],
    echart3: [],
    cpu: [],
    time: []
  },
  database: {
    name: "",
    engine: "",
    version: "",
    size: "",
    kind: "",
  },
  interval_ms: 5*1000,
  str: "string",
  range_start: 25.0,
  range_end: 100.0,
});

export function useState(): { state: State; setState: any } {
  return { state, setState };
}
