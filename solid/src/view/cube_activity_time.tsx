import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import {
  createMemo,
  createResource,
  getOwner,
  mergeProps,
  Show,
} from "solid-js";
import { datazoomEventHandler, listColors } from "../state";
import {
  arrange,
  distinct,
  fixedOrder,
  map,
  pivotWider,
  tidy,
} from "@tidyjs/tidy";
import { ILegend } from "./cube_activity";
import { queryCubeIfLive } from "../http";
import { truncateString } from "../util";
import moment from "moment-timezone";

interface PropsLegend {
  legend: ILegend;
}

export function CubeDimensionTime(props: PropsLegend) {
  const { state, setState } = contextState();
  setState("api", "needDataFor", ["cube_time"]);
  const changed = createMemo((changeCount: number) => {
    // state.timeframe_ms; // handled by createEffect locally
    state.database_instance.dbidentifier;
    state.cubeActivity.uiLegend;
    state.cubeActivity.uiDimension1;
    state.cubeActivity.uiFilter1;
    state.cubeActivity.uiFilter1Value;
    console.log("changed", changeCount);
    return changeCount + 1;
  }, 0);

  const [resourceChanged] = createResource(changed, () => {
    return queryCubeIfLive(state, setState);
  });

  const timezone = moment.tz.guess();
  const timezoneAbbreviation = moment.tz(moment(), timezone).format("z");
  const timeFormat = "HH:mm:ss [GMT]Z ";

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    // highlight: (event: any) => { console.log("Chart Highlight", event); },
    datazoom: datazoomEventHandler,
  };

  const base = {
    color: listColors.map((item) => item.hex),
    animation: false,
    grid: {
      left: 0,
      right: 0,
      top: 10,
      bottom: 0,
      containLabel: true,
    },
    tooltip: {
      trigger: "axis",
      formatter: function (
        params: {
          color: any;
          seriesName: string | number;
          value: { [x: string]: any };
        }[],
      ) {
        const createTooltipRow = (item: {
          color: any;
          seriesName: string | number;
          value: { [x: string]: any };
        }) => {
          const colorDot = `<span style="background-color: ${item.color};" class="inline-block w-3 h-3 rounded-full mr-2"></span>`;

          const truncatedSeriesName = truncateString(
            item.seriesName.toString(),
            80,
          );
          return `
                    <div class="flex items-center mb-1 text-gray-800">
                        ${colorDot}
                        <div class="flex-1 text-left font-medium">${truncatedSeriesName}</div>
                        <div class="ml-auto text-right font-semibold">
                            ${item.value[item.seriesName] || "-"}
                        </div>
                    </div>`;
        };

        return params.map(createTooltipRow).join("");
      },
      axisPointer: {
        type: "cross",
        animation: false,
        label: {
          backgroundColor: "#505765",
        },
      },
    },
    xAxis: {
      type: "category", // NOTE: this isn't "time" because we need to stack the bar chats below.
      axisPointer: {
        label: {
          formatter: function (pointer: { value: string }) {
            let timestamp = parseInt(pointer.value, 10);
            let date = moment(timestamp);
            return date.format(timeFormat) + "(" + timezoneAbbreviation + ")";
          },
        },
      },
      axisLabel: {
        formatter: function (value: string) {
          let timestamp = parseInt(value, 10);
          let date = moment(timestamp);
          return (
            date.format(timeFormat).replace(/ /, "\n") +
            "(" +
            timezoneAbbreviation +
            ")"
          );
        },
      },
    },
    yAxis: {
      type: "value",
    },
    // legend: {
    //   // selectedMode: true,
    //   orient: "vertical",
    //   left: 0,
    //   // top: 70,
    //   // bottom: 20,
    //   textStyle: { color: true },
    // },
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
      // slice(0, 15),
      map((val) => val.out),
    ) as string[];
  });

  return (
    <section class="flex flex-col">
      <section class="h-[28rem]">
        <Show
          when={JSON.stringify(resourceChanged) + JSON.stringify(props.legend)}
          keyed
        >
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
                // emphasis: { focus: "series", },
              })),
              // { label: { show: true, formatter: (params: { value: number }) => { //     return `val1: ${params.value.wait_event_name}: ${params.value.value}`; }, }, },
              // { name: "vCPUs", type: "line", data: [20, 20, 20, 20, 20], markLine: { data: [{ type: "average", name: "Avg" }], },
              // },
              dataZoom: [
                {
                  show: false,
                  //   realtime: true,
                  start: state.range_begin,
                  end: state.range_end,
                  //   xAxisIndex: [0, 1],
                },
                //
                // {
                //   type: "inside",
                //   realtime: true,
                //   start: state.range_begin,
                //   end: state.range_end,
                //   // xAxisIndex: [0, 1],
                // },
              ],
            })}
            eventHandlers={eventHandlers}
          />
        </Show>
      </section>
      <div class="self-end mt-3 text-xs text-neutral-500">
        Number of samples: {dataset().length}
      </div>
    </section>
  );
}
