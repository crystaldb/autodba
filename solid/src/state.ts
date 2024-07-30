import { createStore } from "solid-js/store";

export type State = {
  data: Record<string, any>;
  str: string;
  range_start: number;
  range_end: number;
};

const [state, setState] = createStore({
  data: {},
  str: "string",
  range_start: 0.0,
  range_end: 100.0,
});

export function useState(): { state: State; setState: any } {
  return { state, setState };
}
