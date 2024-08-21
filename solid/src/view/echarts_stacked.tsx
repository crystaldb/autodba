import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { createSignal } from "solid-js";
import { datazoom } from "../event_echarts";
import { State } from "../state";

export function EchartsStacked(props: { class?: string; data: number[][] }) {
  let ref: import("@solid-primitives/refs").Ref<HTMLDivElement>;
  const { state, setState } = contextState();

  const grid = {
    left: 180,
    right: 0,
    top: 10,
    bottom: 20,
  };

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    // highlight: (event: any) => {
    //   console.log("Chart Highlight", event);
    // },
    datazoom: datazoom.bind(null, setState, state),
  };

  const [option] = createSignal(() => {
    // There should not be negative values in rawData
    const rawData: State["data"]["echart1"] = props.data;
    const totalData: number[] = [];
    for (let i = 0; i < rawData[0].length; ++i) {
      let sum = 0;
      for (let j = 0; j < rawData.length; ++j) {
        sum += rawData[j][i];
      }
      totalData.push(sum);
    }
    const series = [
      "CPU",
      "Client:ClientRead",
      "Lock:tuple",
      "LWLock:WALWrite",
      "Lock:transactionid",
    ].map((name, sid) => {
      return {
        name,
        type: "bar",
        stack: "total",
        barWidth: "60%",
        label: {
          show: true,
          formatter: (params: { value: number }) => Math.round(params.value),
          // Math.round(params.value * 1000) / 10 + "%",
        },
        data: rawData[sid].map(
          (d, did) => (totalData[did] <= 0 ? 0 : d),
          // totalData[did] <= 0 ? 0 : d / totalData[did]
        ),
      };
    });
    series.push({
      name: "vCPUs",
      type: "line",
      data: [20, 20, 20, 20, 20],
      // @ts-expect-error markLine is not defined for some reason
      markLine: {
        data: [{ type: "average", name: "Avg" }],
      },
    });

    return {
      legend: {
        selectedMode: true,
        orient: "vertical",
        left: 0,
        top: 70,
        bottom: 20,
        textStyle: {
          color: true,
        },
      },
      grid,
      yAxis: {
        type: "value",
      },
      xAxis: {
        type: "category",
        data: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
      },
      series,
      dataZoom: [
        {
          show: false,
          realtime: true,
          start: state.range_start,
          end: state.range_end,
        },
        {
          type: "inside",
          show: true,
          realtime: true,
          start: state.range_start,
          end: state.range_end,
        },
      ],
    };
  });

  return (
    <div class={`${props.class}`}>
      <EChartsAutoSize
        // @ts-expect-error
        option={option()()}
        eventHandlers={eventHandlers}
        ref={ref}
      />
    </div>
  );
}