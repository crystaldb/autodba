import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { createMemo, mergeProps, Show } from "solid-js";
import { datazoomEventHandler, listColorsHex } from "../state";
import {
  arrange,
  distinct,
  fixedOrder,
  map,
  pivotWider,
  slice,
  tidy,
} from "@tidyjs/tidy";
import { ILegend } from "./cube_activity";

interface PropsLegend {
  legend: ILegend;
}

export function CubeDimensionTime(props: PropsLegend) {
  const { state, setState } = contextState();

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    // highlight: (event: any) => { console.log("Chart Highlight", event); },
    datazoom: datazoomEventHandler.bind(null, setState, state),
  };

  const base = {
    color: listColorsHex,
    animation: false,
    grid: {
      left: 0,
      right: 0,
      top: 10,
      bottom: 25 + 60,
      containLabel: true,
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
      type: "category",
    },
    yAxis: {
      type: "value",
    },
    // legend: { selectedMode: true, orient: "vertical", left: 0, top: 70, bottom: 20, textStyle: { color: true, }, },
  };

  const dataset = createMemo(() => {
    return tidy(
      state.cubeActivity.cubeData,
      (rows) =>
        Array.prototype.concat(
          ...rows.map((row) =>
            row.values.map((val) => ({
              // timestamp: val.timestamp, [row.metric[state.cubeActivity.uiLegend]]: val.value,
              ...row.metric,
              ...val,
            })),
          ),
        ),
      pivotWider({
        namesFrom: state.cubeActivity.uiLegend,
        valuesFrom: "value",
      }),
    );
  });

  const legendDistinct = createMemo<string[]>(() => {
    return tidy(
      state.cubeActivity.cubeData,
      map((row) => ({
        out: row.metric[state.cubeActivity.uiLegend],
      })),
      distinct(({ out }) => out),
      arrange(["out"]),
      arrange([
        // move CPU to the end of the list iff it exists
        fixedOrder((row) => row.out, ["CPU"], { position: "end" }),
      ]),
      slice(0, 15),
      map((val) => val.out),
    ) as string[];
  });

  return (
    <>
      <section class="h-64 min-w-128">
        <Show when={state.cubeActivity.uiLegend} keyed>
          <EChartsAutoSize
            // @ts-expect-error
            option={mergeProps(base, {
              dataset: {
                dimensions: ["timestamp", ...legendDistinct()],
                source: dataset(),
              },
              series: legendDistinct().map(() => ({
                type: "bar",
                barWidth: "50%",
                stack: "time",
                emphasis: {
                  focus: "series",
                },
              })),
              // { label: { show: true, formatter: (params: { value: number }) => { //     return `val1: ${params.value.wait_event_name}: ${params.value.value}`; }, }, },
              // { name: "vCPUs", type: "line", data: [20, 20, 20, 20, 20], markLine: { data: [{ type: "average", name: "Avg" }], },
              // },
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
          />
        </Show>
      </section>
      <details class="mt-20 text-gray-500 text-xs w-full">
        <summary class="text-right">dataset is</summary>{" "}
        {JSON.stringify(dataset(), null, 2)}
      </details>
    </>
  );
}
