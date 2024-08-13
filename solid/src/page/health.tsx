import { contextState } from "../context_state";
import { JSX } from "solid-js";
import { EchartsGauge1 } from "../view/echarts_gauge_1";
import { EchartsTimeseries1 } from "../view/echarts_timeseries_1";
import { AiOutlineInfoCircle } from "solid-icons/ai";
import { Popover } from "solid-simple-popover";
import { flip } from "@floating-ui/dom";

export function PageHealth(props: any) {
  const page = "pageHealth";
  const { state } = contextState();

  return (
    <>
      <section class="grid gap-4 grid-cols-1 xs:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
        <MetricColumn>
          {header("info-Connections", "Connections")}
          <div class="flex flex-col h-28">
            <EchartsGauge1 data={state.data.cpu} />
          </div>
          <div class="flex flex-col h-64">
            <EchartsTimeseries1 time={state.data.time} data={state.data.cpu} />
          </div>
        </MetricColumn>

        <MetricColumn>
          {header("info-Processes", "Processes")}
          <div class="flex flex-col h-28">
            <EchartsGauge1 data={state.data.cpu.map((x) => x - 10)} />
          </div>
          <div class="flex flex-col h-64">
            <EchartsTimeseries1
              time={state.data.time}
              data={state.data.cpu.map((x) => x - 10)}
            />
          </div>
        </MetricColumn>

        <MetricColumn>
          {header("info-Memory", "Memory")}
          <div class="flex flex-col h-28">
            <EchartsGauge1 data={state.data.cpu.map((x) => x + 10)} />
          </div>
          <div class="flex flex-col h-64">
            <EchartsTimeseries1
              time={state.data.time}
              data={state.data.cpu.map((x) => x + 10)}
            />
          </div>
        </MetricColumn>

        <MetricColumn>
          {header("info-Disk", "Disk")}
          <div class="flex flex-col h-28">
            <EchartsGauge1 data={state.data.cpu.map((x) => x - 20)} />
          </div>
          <div class="flex flex-col h-64">
            <EchartsTimeseries1
              time={state.data.time}
              data={state.data.cpu.map((x) => x - 20)}
            />
          </div>
        </MetricColumn>
      </section>
    </>
  );
}

function MetricColumn(props: {
  children:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}) {
  return (
    <section class="flex flex-col w-64 gap-y-4 p-4 rounded-xl bg-neutral-100 dark:bg-neutral-800">
      {props.children}
    </section>
  );
}

function header(
  id: string,
  text: string,
  info: string = "TODO replace with helpful info"
) {
  return (
    <div class="flex justify-between relative">
      <h2>{text}</h2>
      <button id={id} class="w-6">
        <AiOutlineInfoCircle size="24" class="text-gray-500" />
      </button>

      <Popover
        triggerElement={`#${id}`}
        autoUpdate
        computePositionOptions={{
          placement: "bottom-start",
          middleware: [flip()],
        }}
        // sameWidth
        dataAttributeName="data-open"
      >
        <div class="floating width60">{info}</div>
      </Popover>
    </div>
  );
}
