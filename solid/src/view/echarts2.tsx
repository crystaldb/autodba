import type { ECharts } from "echarts-solid";
import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { batch, createSignal, onMount } from "solid-js";
import { datazoom } from "../event_echarts";

export function Echarts2(props: { dataA: any; dataB: any; dataC: any }) {
  const { state, setState } = contextState();

  const [option] = createSignal(() => {
    // NOTE: force update. Without this, the graph does not update
    let a = props.dataA.length;
    let b = props.dataB.length;
    let c = props.dataC.length;
    // END NOTE
    return {
      title: {
        text: "Rainfall and Flow Relationship",
        left: "center",
      },
      grid: {
        bottom: 80,
      },
      toolbox: {
        feature: {
          dataZoom: {
            yAxisIndex: "none",
          },
          restore: {},
          saveAsImage: {},
        },
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
      legend: {
        data: ["Flow", "Rainfall"],
        left: 10,
      },
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
      xAxis: [
        {
          type: "category",
          boundaryGap: false,
          axisLine: { onZero: false },
          data: props.dataA,
        },
      ],
      yAxis: [
        {
          name: "Flow(m³/s)",
          type: "value",
        },
        {
          name: "Rainfall(mm)",
          nameLocation: "start",
          alignTicks: true,
          type: "value",
          inverse: true,
        },
      ],
      series: [
        {
          name: "Flow",
          type: "line",
          areaStyle: {},
          lineStyle: {
            width: 1,
          },
          emphasis: {
            focus: "series",
          },
          markArea: {
            silent: true,
            itemStyle: {
              opacity: 0.3,
            },
            data: [
              [
                {
                  xAxis: "2009/9/12\n7:00",
                },
                {
                  xAxis: "2009/9/22\n7:00",
                },
              ],
            ],
          },
          data: props.dataB,
        },
        {
          name: "Rainfall",
          type: "line",
          yAxisIndex: 1,
          areaStyle: {},
          lineStyle: {
            width: 1,
          },
          emphasis: {
            focus: "series",
          },
          markArea: {
            silent: true,
            itemStyle: {
              opacity: 0.3,
            },
            data: [
              [
                {
                  xAxis: "2009/9/10\n7:00",
                },
                {
                  xAxis: "2009/9/20\n7:00",
                },
              ],
            ],
          },
          data: props.dataC,
        },
      ],
    };
  });

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    highlight: (event: any) => {
      // console.log("Chart2 Highlight", event);
    },
    timelinechanged: (event: any) => {
      console.log("Chart2 Timeline Changed", event);
    },
    datarangeselected: (event: any) => {
      console.log("Chart2 Data Range Selected", event);
    },
    datazoom: datazoom.bind(null, setState, state),
    dataviewchanged: (event: any) => {
      console.log("Chart2 Data View Changed", event);
    },
  };

  // onMount(() => {
  //   batch(() => {
  //     setState("range_start", 25);
  //     setState("range_end", 85);
  //   });
  // });

  return (
    <>
      <EChartsAutoSize
        // @ts-expect-error
        option={option()()}
        eventHandlers={eventHandlers}
        class="bg-white dark:bg-black"
      />
    </>
  );
}
