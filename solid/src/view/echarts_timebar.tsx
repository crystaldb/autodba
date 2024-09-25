import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { mergeProps, Show } from "solid-js";
import { datazoomEventHandler, getTimeAtPercentage } from "../state";
import moment from "moment-timezone";

interface IEchartsTimebarProps {
  class?: string;
}

export function EchartsTimebar(props: IEchartsTimebarProps) {
  const { state } = contextState();
  const dateZero = +new Date();
  const base = {
    grid: {
      left: 0,
      right: 7,
    },
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

  const timezone = moment.tz.guess();
  const timezoneAbbreviation = moment.tz(moment(), timezone).format("z");
  const timeFormat = "HH:mm:ss [GMT]Z ";

  const datasource = (): number[] => {
    const timeEnd = state.server_now || dateZero;
    const timeRef = { timeframe_ms: state.timeframe_ms, server_now: timeEnd };
    const windowBegin = Math.floor(
      getTimeAtPercentage(timeRef, state.range_begin),
    );
    return [
      timeEnd - state.timeframe_ms,
      windowBegin,
      Math.max(
        windowBegin + 1,
        Math.ceil(getTimeAtPercentage(timeRef, state.range_end)),
      ),
      timeEnd,
    ];
  };

  return (
    <div class={`relative ${props.class}`}>
      <EChartsAutoSize
        // @ts-expect-error eCharts types don't seem correct, so suppress TS error
        option={mergeProps(base, {
          dataset: { source: datasource() },
          dataZoom: [
            {
              type: "slider",
              show: true,
              realtime: false,
              start: state.range_begin,
              end: state.range_end,
            },
          ],
        })}
        eventHandlers={eventHandlers}
      />
      <section class="text-neutral-500">
        <Show
          when={
            datasource()[0] !== datasource()[1] ||
            datasource()[2] !== datasource()[3]
          }
        >
          <p class="absolute isolate -z-10 -top-4 inset-x-0 text-sm px-1 flex bg-inherit">
            <span
              style={{ "margin-left": state.range_begin + "%" }}
              class="absolute z-10 bg-zinc-100 dark:bg-zinc-900 p-0.5 rounded"
            >
              {moment(datasource()[1]).format(timeFormat).split(/ /)[0]}
            </span>
            <span
              style={{
                "margin-left": "calc(" + state.range_end + "% - 3rem)",
                opacity:
                  "calc( 2 * " +
                  (state.range_end - state.range_begin) / 100 +
                  " )",
              }}
              class="-z-10 bg-zinc-100 dark:bg-zinc-900 p-0.5 rounded"
            >
              {moment(datasource()[2]).format(timeFormat).split(/ /)[0]}
            </span>
          </p>
        </Show>
        <p class="absolute -z-10 top-4 inset-x-0 text-sm px-2.5 flex justify-between">
          <span class="bg-zinc-100 dark:bg-zinc-900 p-0.5 rounded">
            {moment(datasource()[0]).format(timeFormat).split(/ /)[0]}
          </span>
          <span class="bg-zinc-100 dark:bg-zinc-900 p-0.5 rounded">
            {moment(datasource()[0]).format(timeFormat).split(/ /)[1] +
              " (" +
              timezoneAbbreviation +
              ")"}
          </span>
          <span class="bg-zinc-100 dark:bg-zinc-900 p-0.5 rounded">
            {moment(datasource()[3]).format(timeFormat).split(/ /)[0]}
          </span>
        </p>
      </section>
    </div>
  );
}
