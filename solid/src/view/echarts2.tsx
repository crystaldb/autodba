import type { ECharts } from "echarts-solid";
import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { createSignal, mergeProps } from "solid-js";
import { datazoomEventHandler } from "../state";

interface PropsEcharts2 {
  title: string;
  metricList: string[];
  data: any[];
  class?: string;
}

export function Echarts2(props: PropsEcharts2) {
  const { state } = contextState();

  const base = {
    grid: {
      // bottom: 75,
      containLabel: true,
    },
    legend: {
      data: props.metricList,
      itemGap: 1,
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
      // type: "category",
      type: "time",
      // boundaryGap: false,
      // axisLine: { onZero: false },
    },
    yAxis: {
      type: "value",
      //   name: "    Requests/sec",
    },
    series: props.metricList.map((metric) => ({
      name: metric,
      type: "line",
      // stack: "Total",
      dimensions: ["time_ms", metric],
      // name: "Requests",
      // // areaStyle: {},
      // lineStyle: { width: 1, },
      // emphasis: { focus: "series", },
      // markArea: {
      //   silent: true, itemStyle: { opacity: 0.3, },
      //   },
    })),
    // title: { text: props.title, left: -5, textStyle: { fontSize: 14, }, },
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
    datazoom: datazoomEventHandler,
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
            // dimensions: ["time_ms", ...props.metricList],
            source: props.data,
            forceSolidRefresh: props.data.length,
          },
          dataZoom: [
            // {
            //   type: "inside",
            //   start: state.range_begin,
            //   end: state.range_end,
            // },
            {
              show: false,
              start: state.range_begin,
              end: state.range_end,
            },
          ],
        })}
        eventHandlers={eventHandlers}
      />
    </div>
  );
}
