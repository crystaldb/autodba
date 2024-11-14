import { graphic } from "echarts";
import { EChartsAutoSize } from "echarts-solid";
import { mergeProps } from "solid-js";
import { contextState } from "../context_state";
import { datazoomEventHandler } from "../state";

interface IProps {
  time: number[];
  data: number[];
}

export function EchartsTimeseries1(props: IProps) {
  let ref: import("@solid-primitives/refs").Ref<HTMLDivElement>;
  const { state } = contextState();

  const base = {
    tooltip: {
      trigger: "axis",
      position: (pt: number[]) => [pt[0], "10%"],
    },
    // title: {
    //   left: "center",
    //   text: "CPU",
    // },
    toolbox: {
      feature: {
        dataZoom: {
          yAxisIndex: "none",
        },
        restore: {},
        saveAsImage: {},
      },
    },
    xAxis: {
      type: "time",
      boundaryGap: false,
    },
    yAxis: {
      type: "value",
      boundaryGap: [0, "100%"],
    },
    series: [
      {
        name: "Fake Data",
        type: "line",
        step: true,
        smooth: false,
        symbol: "none",
        itemStyle: {
          // color: "#FFAB91",
          color: "fuchsia",
        },
        areaStyle: {
          color: new graphic.LinearGradient(0, 0, 0, 1, [
            {
              offset: 0,
              color: "fuchsia",
            },
            {
              offset: 1,
              color: "transparent",
            },
          ]),
        },
      },
    ],
    // RESPONSIVE CONFIG BELOW
    // dataZoom: [
    //   {
    //     type: "inside",
    //     start: state.range_begin,
    //     end: state.range_end,
    //   },
    //   {
    //     start: state.range_begin,
    //     end: state.range_end,
    //   },
    // ],
    // dataset: { source: props.data },
  };

  const eventHandlers = {
    click: (event: Event) => {
      console.log("Chart is clicked!", event);
    },
    highlight: (event: Event) => {
      console.log("Chart Highlight", event);
    },
    datazoom: datazoomEventHandler,
  };

  return (
    <>
      <EChartsAutoSize
        // @ts-expect-error Type for eCharts option may be incorrect or incomplete. It complains abou xAxis.
        option={mergeProps(base, {
          dataset: { source: props.data.map((d, i) => [props.time[i], d]) },
          dataZoom: [
            {
              type: "inside",
              start: state.range_begin,
              end: state.range_end,
            },
            {
              start: state.range_begin,
              end: state.range_end,
            },
          ],
        })}
        eventHandlers={eventHandlers}
        ref={ref}
        class=""
      />
    </>
  );
}
