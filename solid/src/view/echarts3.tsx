import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { batch, createSignal } from "solid-js";
import { datazoom } from "../state";

export function Echarts3(this: any) {
  let ref: import("@solid-primitives/refs").Ref<HTMLDivElement>;
  const { state, setState } = contextState();

  const [option] = createSignal(() => ({
    xAxis: {
      type: "category",
      data: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
    },
    yAxis: {
      type: "value",
    },
    series: [
      {
        data: [150, 230, 224, 218, 135, 147, 260],
        type: "line",
      },
    ],
    dataZoom: [
      {
        show: true,
        realtime: true,
        start: state.range_start,
        end: state.range_end,
      },
      {
        type: "inside",
        realtime: true,
        start: state.range_start,
        end: state.range_end,
      },
    ],
  }));

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    highlight: (event: any) => {
      console.log("Chart Highlight", event);
    },
    datazoom: datazoom.bind(null, setState, state),
  };

  return (
    <>
      <EChartsAutoSize
        // @ts-expect-error
        option={option()()}
        eventHandlers={eventHandlers}
        ref={ref}
        class=""
      />
    </>
  );
}
