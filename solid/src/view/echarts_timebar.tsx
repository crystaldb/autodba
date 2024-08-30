import { ECharts } from "echarts-solid";
import { contextState } from "../context_state";
import { mergeProps } from "solid-js";
import { datazoomEventHandler } from "../state";

interface IEchartsTimebarProps {
  class?: string;
}

export function EchartsTimebar(props: IEchartsTimebarProps) {
  const { state, setState } = contextState();
  const base = {
    xAxis: [
      {
        type: "category",
      },
    ],
    yAxis: [
      {
        type: "value",
      },
    ],
  };

  const eventHandlers = {
    // click: (event: any) => { console.log("Chart is clicked!", event); },
    // highlight: (event: any) => { console.log("Chart Highlight", event); },
    datazoom: datazoomEventHandler.bind(null, setState, state),
  };

  return (
    <ECharts
      // @ts-expect-error suppress complaint about `type: "gauge"`
      option={mergeProps(base, {
        dataZoom: [
          {
            show: true,
            realtime: true,
            start: state.range_start,
            end: state.range_end,
            xAxisIndex: [0, 1],
          },
          {
            type: "inside",
            realtime: true,
            start: state.range_start,
            end: state.range_end,
            xAxisIndex: [0, 1],
          },
        ],
      })}
      eventHandlers={eventHandlers}
      class={props.class}
    />
  );
}
