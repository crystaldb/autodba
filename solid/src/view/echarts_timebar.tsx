import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { mergeProps } from "solid-js";
import { datazoomEventHandler } from "../state";

interface IEchartsTimebarProps {
  class?: string;
}

export function EchartsTimebar(props: IEchartsTimebarProps) {
  const { state } = contextState();
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
    datazoom: datazoomEventHandler,
  };

  return (
    <div class={props.class}>
      <EChartsAutoSize
        // @ts-expect-error eCharts types don't seem correct, so suppress TS error
        option={mergeProps(base, {
          dataZoom: [
            {
              show: true,
              type: "slider",
              realtime: false,
              start: state.range_begin,
              end: state.range_end,
            },
          ],
        })}
        eventHandlers={eventHandlers}
        class="border border-red-500"
      />
    </div>
  );
}
