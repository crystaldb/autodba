import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { mergeProps } from "solid-js";
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
      containLabel: true,
      top: 20 + props.metricList.length * 13, // space for the legend
      bottom: 0,
      right: 0,
      left: 0,
    },
    legend: {
      data: props.metricList,
      itemGap: 1,
      left: 0,
      textStyle: {
        color: "gray",
      },
    },
    tooltip: {
      trigger: "axis",
      valueFormatter: (value: number) => value?.toFixed(1),
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
      axisLabel: {
        rotate: 45,
      },
    },
    yAxis: {
      type: "value",
    },
    series: props.metricList.map((metric) => ({
      name: metric,
      type: "line",
      // stack: "Total",
      dimensions: ["time_ms", metric],
      // // areaStyle: {},
      // lineStyle: { width: 1, },
      // emphasis: { focus: "series", },
      // markArea: {
      //   silent: true, itemStyle: { opacity: 0.3, },
      //   },
    })),
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
            source: props.data,
            // dimensions: ["time_ms", ...props.metricList],
            // forceSolidRefresh: props.data.length,
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
