import { batch } from "solid-js";

export const datazoom = (
  setState: (arg0: string, arg1: any) => void,
  stateFn: any,
  event: any
) => {
  console.log("Chart2 Data Zoom", event);
  batch(() => {
    setState("range_start", event.start || event.batch?.at(0)?.start);
    setState("range_end", event.end || event.batch?.at(0)?.end);
  });
};
