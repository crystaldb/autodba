import { createStore } from "solid-js/store";

export type State = {
  // data: Record<string, any>;
  data: {
    echart1: number[][];
    echart2a: any[];
    echart2b: any[];
    echart2c: any[];
    echart3: any[];
  };
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
  },
  str: "string",
  range_start: 50.0,
  range_end: 100.0,
});

export function useState(): { state: State; setState: any } {
  return { state, setState };
}
