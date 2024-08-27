import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { createMemo, createSignal, For, JSX, Show } from "solid-js";
import {
  datazoom,
  DimensionField,
  listWaitsColorsText,
  DimensionName,
  listDimensionTabNames,
  CubeData,
  listWaitsColorsBg,
} from "../state";
import {
  arrange,
  distinct,
  filter,
  first,
  fixedOrder,
  groupBy,
  map,
  sum,
  summarize,
  tidy,
} from "@tidyjs/tidy";

type ILegend = {
  item: string;
  colorText: string;
  colorBg: string;
}[];

export function CubeActivity() {
  const { state } = contextState();

  const cubeData = createMemo<CubeData>(() => {
    return tidy(
      state.cubeActivity.cubeData,
      filter(
        (row) =>
          !!row.metric[state.cubeActivity.uiLegend] &&
          (row.metric[state.cubeActivity.uiDimension1] === DimensionName.time ||
            !!row.metric[state.cubeActivity.uiDimension1]),
      ),
    );
  });

  const distinctLegend = createMemo((): ILegend => {
    return tidy(
      cubeData(),
      distinct((row) => row.metric[state.cubeActivity.uiLegend]),
      map((row) => ({
        item: row.metric[state.cubeActivity.uiLegend],
      })),
      arrange([
        // move CPU to the end of the list iff it exists
        fixedOrder((row) => row.item, ["CPU"], { position: "end" }),
      ]),
      filter(({ item }) => !!item),
      map((item, index) => ({
        item: item.item!,
        colorText: listWaitsColorsText[index] || "",
        colorBg: listWaitsColorsBg[index] || "",
      })),
    );
  });

  return (
    <>
      <section class="flex flex-col md:flex-row items-start gap-4">
        <section class="flex flex-col gap-4 ps-4 border-s border-neutral-300 dark:border-neutral-700">
          <div class="flex flex-wrap gap-x-3 text-sm">
            <label class="font-medium">Slice/Color by</label>
            <SelectSliceBy dimension={DimensionField.uiLegend} />
          </div>
          <Legend legend={distinctLegend()} />
        </section>
        <section class="flex flex-col gap-5">
          <section class="flex flex-col gap-y-5">
            <div class="flex gap-3 text-sm">
              <h2 class="font-medium">Dimensions</h2>
              <DimensionTabs dimension="uiDimension1" cubeData={cubeData} />
            </div>
          </section>
          <Dimension1 cubeData={cubeData} legend={distinctLegend()} />
        </section>
      </section>

      <details>
        <summary class="text-gray-500">debug0</summary>
        <section class="text-gray-500">
          <div>
            legend & dimension{">>>"}
            {state.cubeActivity.uiLegend}::
            {state.cubeActivity.uiDimension1}:::{state.cubeActivity.uiFilter1}
          </div>
          <div>legend is {JSON.stringify(distinctLegend())}</div>
        </section>
      </details>

      <details>
        <summary class="text-gray-500">debug2</summary>
        <div>
          <span>URL</span> /api/v1/activity?database_list={state.database.name}
          &start={state.timeframe_start_ms}
          &end={state.timeframe_end_ms}
          &step={state.interval_ms}ms&legend=
          <span class="text-green-500">{state.cubeActivity.uiLegend}</span>
          &dim=
          <span class="text-green-500">{state.cubeActivity.uiDimension1}</span>
          &filterdim=
          <span class="text-green-500">{state.cubeActivity.uiFilter1}</span>
          &filterdimselected=
          <span class="text-green-500">
            {encodeURIComponent(state.cubeActivity.uiFilter1Value || "")}
          </span>
        </div>

        <pre class="text-xs whitespace-pre-wrap break-works text-neutral-500 max-w-28 dark:text-neutral-400">
          Cube data
          <br />
          {JSON.stringify(state.cubeActivity, null, 2)}
        </pre>
      </details>
    </>
  );
}

interface PropsLegend {
  legend: ILegend;
}

function Legend(props: PropsLegend) {
  const { state, setState } = contextState();
  return (
    <section class="flex flex-col gap-4">
      <For each={props.legend}>
        {(item, index) => (
          <label>
            <span class={item.colorText}>{item.item}</span>
          </label>
        )}
      </For>
    </section>
  );
}

interface IDimension1 {
  cubeData: () => CubeData;
  class?: string;
  legend: ILegend;
}

function Dimension1(props: IDimension1) {
  const { state } = contextState();
  return (
    <>
      <Show
        when={state.cubeActivity.uiDimension1 === DimensionName.time}
        fallback={
          <DimensionView cubeData={props.cubeData} legend={props.legend} />
        }
      >
        <div class={`${props.class}`}>
          <DimensionTime />
        </div>
      </Show>
    </>
  );
}

function DimensionTime() {
  let ref: import("@solid-primitives/refs").Ref<HTMLDivElement>;
  const { state, setState } = contextState();

  const grid = {
    left: 180,
    right: 0,
    top: 10,
    bottom: 20,
  };

  const [option] = createSignal(() => {
    let series = [
      {
        name: "CPU",
        type: "bar",
        stack: "total",
        barWidth: "60%",
        label: {
          show: true,
          formatter: (params: { value: number }) => Math.round(params.value),
        },
        data: [20, 10, 30, 10, 25],
      },
      {
        name: "vCPUs",
        type: "line",
        data: [20, 20, 20, 20, 20],
        markLine: {
          data: [{ type: "average", name: "Avg" }],
        },
      },
    ];
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
        type: "time",
        data: state.cubeActivity.arrTime,
      },
      series,
      dataZoom: [
        {
          show: true,
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

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    // highlight: (event: any) => { console.log("Chart Highlight", event); },
    datazoom: datazoom.bind(null, setState, state),
  };

  return (
    <div class="h-64 min-w-128">
      <EChartsAutoSize
        // @ts-expect-error
        option={option()()}
        eventHandlers={eventHandlers}
        ref={ref}
      />
    </div>
  );
}

interface IDimensionView {
  cubeData: () => CubeData;
  legend: ILegend;
}

function DimensionView(props: IDimensionView) {
  const { state } = contextState();

  const distinctDimension1 = (): {
    dimensionValue: string;
    total: number;
    records: { metric: { [key: string]: string }; values: { value: any }[] }[];
  }[] => {
    if (state.cubeActivity.uiDimension1 === DimensionName.time) {
      return [];
    }
    return tidy(
      props.cubeData(),
      filter(
        (d) =>
          !!d.metric[state.cubeActivity.uiDimension1] &&
          (d.metric[state.cubeActivity.uiDimension1] === DimensionName.time ||
            !!d.metric[state.cubeActivity.uiDimension1]),
      ),
      groupBy(
        (d) => d.metric[state.cubeActivity.uiDimension1],
        [
          summarize({
            dimensionValue: first(
              (d: {
                metric: Record<string, string>;
                values: { value: number }[];
              }) => {
                return d.metric[state.cubeActivity.uiDimension1];
              },
            ),
            total: sum(
              (d: { values: { value: number }[] }) => d.values[0].value,
            ),
            records: (d) => d,
          }),
        ],
      ),
    );
  };

  return (
    <section class="flex flex-col gap-4">
      <For each={distinctDimension1()}>
        {({ total, dimensionValue, records }) => (
          <DimensionRow
            len={total}
            txt={dimensionValue}
            records={records}
            legend={props.legend}
          />
        )}
      </For>
      <details>
        <summary class="text-gray-500">debug1</summary>
        <div class="word-break whitespace-pre">
          {JSON.stringify(distinctDimension1(), null, 2)}
        </div>
        <For each={distinctDimension1()}>
          {(value) => <div>dimension1 is {JSON.stringify(value)}</div>}
        </For>
      </details>
    </section>
  );
}

interface IDimensionRow {
  legend: ILegend;
  records: any;
  len: number;
  txt:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}

function DimensionRow(props: IDimensionRow) {
  const { state } = contextState();
  return (
    <section class="flex items-center">
      <div class="w-48 xs:w-64 flex flex-row">
        <For each={props.records}>
          {(record) => (
            <DimensionRowPart
              len={record.values[0].value}
              txt={record.metric[state.cubeActivity.uiLegend]}
              legend={props.legend}
            />
          )}
        </For>
      </div>
      <div class="grow">{props.txt}</div>
    </section>
  );
}

interface IDimensionRowPart {
  len: number;
  txt: string & {};
  legend: ILegend;
}

function DimensionRowPart(props: IDimensionRowPart) {
  let css = "";
  for (let i = 0; i < props.legend.length; i++) {
    if (props.legend[i].item === props.txt) {
      css = props.legend[i].colorBg;
    }
  }
  return (
    <div
      style={{ width: `${props.len * 10}%` }}
      class={`rounded cursor-default ${css}`}
      title={props.txt}
    >
      {props.len.toFixed(1)}
    </div>
  );
}

interface IDimensionTabs {
  dimension: "uiDimension1";
  cubeData: () => CubeData;
}

function DimensionTabs(props: IDimensionTabs) {
  const { state, setState } = contextState();

  return (
    <section class="flex flex-col">
      <section class="flex flex-wrap gap-3">
        <Tab
          value={DimensionName.time}
          txt="Time"
          selected={state.cubeActivity[props.dimension] === DimensionName.time}
        />
        <For each={listDimensionTabNames()}>
          {(value) => (
            <Tab
              value={value[0]}
              txt={`${value[1]}`}
              selected={state.cubeActivity[props.dimension] === value[0]}
            />
          )}
        </For>
        <section class="ms-6 flex flex-wrap items-center gap-x-3 text-sm">
          <label class="font-medium">Filter by</label>
          <SelectSliceBy
            dimension={DimensionField.uiFilter1}
            list={[["none", "No filter"], ...listDimensionTabNames()]}
          />
        </section>
      </section>
      <Show when={state.cubeActivity.uiFilter1 !== DimensionName.none}>
        <div class="self-end flex flex-wrap items-center gap-x-3 text-sm">
          <button
            class="hover:underline underline-offset-4 me-4"
            onClick={() => {
              setState("cubeActivity", "uiFilter1Value", "");
            }}
          >
            clear
          </button>
          <SelectSliceBy
            dimension="uiFilter1Value"
            list={listFor(DimensionField.uiFilter1, props.cubeData)}
            class="grow max-w-screen-sm"
          />
        </div>
      </Show>
    </section>
  );
}

function Tab(props: { value: string; txt: string; selected: boolean }) {
  const { setState } = contextState();
  return (
    <button
      value={props.value}
      class="px-1.5 border-x-2"
      classList={{
        "text-black dark:text-white border-fuchsia-500 dark:border-fuchsia-500 rounded":
          props.selected,
        "text-neutral-600 dark:text-neutral-400 border-transparent bg-neutral-100 dark:bg-neutral-800":
          !props.selected,
      }}
      onClick={() => setState("cubeActivity", "uiDimension1", props.value)}
    >
      {props.txt}
    </button>
  );
}

function SelectSliceBy(props: {
  dimension: DimensionField | "uiFilter1Value";
  class?: string;
  list?: string[][];
}) {
  const { state, setState } = contextState();

  return (
    <select
      onChange={(event) => {
        const value = event.target.value;
        setState("cubeActivity", props.dimension, value);
      }}
      class={`bg-transparent border-x-2 border-fuchsia-500 rounded text-fuchsia-500 ps-2 pe-2 hover:border-gray-400 focus:outline-none ${props.class}`}
    >
      <For each={props.list || listDimensionTabNames()}>
        {(value) => (
          <Option
            value={value[0]}
            txt={value[1]}
            selected={state.cubeActivity[props.dimension] === value[0]}
          />
        )}
      </For>
    </select>
  );
}

function Option(props: { value: string; txt: string; selected: boolean }) {
  return (
    <option
      value={props.value}
      selected={props.selected || undefined}
      class="appearance-none bg-neutral-100 dark:bg-neutral-800"
    >
      {props.txt}
    </option>
  );
}

function listFor(
  dimensionField: DimensionField,
  cubeData: () => CubeData,
): [string, string][] {
  if (dimensionField !== DimensionField.uiFilter1) return [];
  const { state } = contextState();
  const dimensionName: DimensionName = state.cubeActivity[dimensionField];

  const input = tidy(
    cubeData(),
    filter((d) => !!d.metric[dimensionName]),
    map((d) => ({ result: d.metric[dimensionName] })),
    distinct((d) => d.result),
    filter(({ result }) => !!result),
  ).map(({ result }) => result!);

  let list: [string, string][] = input.map((x) => [x, x]);

  list.unshift(["", "no filter"]);
  return list;
}
