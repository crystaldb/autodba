import { ECharts, EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { mergeProps, Show } from "solid-js";
import { datazoomEventHandler } from "../state";
import { isLiveQueryCube } from "../http";

let debugZero = +new Date();

interface IEchartsTimebarProps {
  class?: string;
}

export function EchartsTimebar(props: IEchartsTimebarProps) {
  const { state, setState } = contextState();
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
    <Show when={state.time_begin_ms && state.time_end_ms}>
      <div class={`flex items-center ${props.class}`}>
        <div class="flex flex-col text-xs text-gray-600 dark:text-gray-400 w-50 shrink-0">
          <p>timeBegin: {state.time_begin_ms - debugZero}</p>
          <p>timeEnd: {state.time_end_ms - debugZero}</p>
          <p>windowBegin: {state.window_begin_ms - debugZero}</p>
          <p>windowEnd: {state.window_end_ms - debugZero}</p>
        </div>
        <div class={`flex flex-col ${props.class}`}>
          <div class="flex border border-green-500 h-6 ms-[5.5rem] me-[5.2rem]">
            <div
              class="bg-yellow-500 h-full"
              style={{
                width: `${
                  ((state.window_begin_ms! - state.time_begin_ms!) /
                    (state.time_end_ms! - state.time_begin_ms!)) *
                  100
                }%`,
              }}
            ></div>
            <div
              class="bg-green-500 h-full"
              style={{
                width: `${
                  ((state.window_end_ms! - state.window_begin_ms!) /
                    (state.time_end_ms! - state.time_begin_ms!)) *
                  100
                }%`,
              }}
            ></div>
            <div
              class="bg-red-500 h-full"
              style={{
                width: `${
                  100 -
                  ((state.window_end_ms! - state.time_begin_ms!) /
                    (state.time_end_ms! - state.time_begin_ms!)) *
                    100
                }%`,
              }}
            ></div>
          </div>
          {/*AutoSize
           */}
          <EChartsAutoSize
            // @ts-expect-error suppress complaint about `type: "gauge"`
            option={mergeProps(base, {
              dataZoom: [
                {
                  show: true,
                  type: "slider",
                  realtime: false,
                  start: state.range_begin,
                  end: state.range_end,
                  // xAxisIndex: [0, 1],
                },
                // {
                //   type: "inside",
                //   realtime: false,
                //   start: state.range_begin,
                //   end: state.range_end,
                //   // xAxisIndex: [0, 1],
                // },
              ],
            })}
            eventHandlers={eventHandlers}
            class="border border-red-500 h-12 max-h-8"
          />
        </div>

        <Show when={isLiveQueryCube(state)}>
          <div class="p-2 rounded-md bg-yellow-200 text-black">LIVE</div>
        </Show>
      </div>
    </Show>
  );
}
