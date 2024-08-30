import type { ECharts } from "echarts-solid";
import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { createSignal, mergeProps } from "solid-js";
import { datazoomEventHandler } from "../state";

export function Echarts2(props: {
  title: string;
  metric: string;
  data: any[];
  class?: string;
}) {
  const { state, setState } = contextState();

  const base = {
    grid: {
      bottom: 75,
    },
    tooltip: {
      trigger: "axis",
      axisPointer: {
        type: "cross",
        animation: false,
        label: {
          backgroundColor: "#505765",
        },
      },
    },
    xAxis: {
      type: "time",
      // boundaryGap: false,
      // axisLine: { onZero: false },
      // data: props.dataA,
    },
    yAxis: {
      type: "value",
      //   name: "    Requests/sec",
    },
    series: [
      {
        type: "line",
        // name: "Requests",
        // // areaStyle: {},
        // lineStyle: { width: 1, },
        // emphasis: { focus: "series", },
        // markArea: {
        //   silent: true, itemStyle: { opacity: 0.3, },
        //   data: [ [ { xAxis: "2009/9/12\n7:00", }, { xAxis: "2009/9/22\n7:00", }, ], ],
        //   },
        // data: props.dataB,
      },
    ],
    // title: { text: props.title, left: -5, textStyle: { fontSize: 14, }, },
    // legend: { data: ["Requests", "Requests 2"], left: 0, bottom: 40, },
  };

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    // highlight: (event: any) => {
    //   console.log("Chart2 Highlight", event);
    // },
    timelinechanged: (event: any) => {
      console.log("Chart2 Timeline Changed", event);
    },
    datarangeselected: (event: any) => {
      console.log("Chart2 Data Range Selected", event);
    },
    datazoom: datazoomEventHandler.bind(null, setState, state),
    dataviewchanged: (event: any) => {
      console.log("Chart2 Data View Changed", event);
    },
  };

  return (
    <div class={props.class}>
      <EChartsAutoSize
        // @ts-expect-error ECharts type is not complete
        option={mergeProps(base, {
          dataset: {
            dimensions: ["time_ms", props.metric],
            forceSolidRefresh: props.data.length,
            source: props.data,
          },
          dataZoom: [
            {
              type: "inside",
              start: state.range_start,
              end: state.range_end,
            },
            {
              start: state.range_start,
              end: state.range_end,
            },
          ],
        })}
        eventHandlers={eventHandlers}
      />
    </div>
  );
}
