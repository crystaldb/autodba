import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { mergeProps } from "solid-js";
import { datazoomEventHandler } from "../state";
import { graphic } from "echarts";

interface IProps {
  time: number[];
  data: number[];
}

export function EchartsTimeseries1(props: IProps) {
  let ref: import("@solid-primitives/refs").Ref<HTMLDivElement>;
  const { state, setState } = contextState();

  const base = {
    tooltip: {
      trigger: "axis",
      position: function (pt: any[]) {
        return [pt[0], "10%"];
      },
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
    //     start: state.range_start,
    //     end: state.range_end,
    //   },
    //   {
    //     start: state.range_start,
    //     end: state.range_end,
    //   },
    // ],
    // dataset: { source: props.data },
  };

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    highlight: (event: any) => {
      console.log("Chart Highlight", event);
    },
    datazoom: datazoomEventHandler.bind(null, setState, state),
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
        ref={ref}
        class=""
      />
    </>
  );
}
