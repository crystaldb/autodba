import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import {
  createEffect,
  createSignal,
  For,
  JSX,
  Match,
  onMount,
  Show,
  Switch,
} from "solid-js";
import { datazoom } from "../event_echarts";
import { listWaits, listWaitsColors } from "../state";

const listDefault = [
  ["waits", "Waits"],
  ["sql", "Sql"],
  ["hosts", "Hosts"],
  ["users", "Users"],
  ["session_types", "Session types"],
  ["applications", "Applications"],
  ["databases", "Databases"],
];

interface ICubeActivity {
  class?: string;
}

export function CubeActivity(props: ICubeActivity) {
  const { state, setState } = contextState();

  return (
    <>
      <section class="flex flex-col md:flex-row items-start gap-4">
        <section class="flex flex-col gap-5">
          <section class="flex flex-col gap-y-5">
            <div class="flex gap-3 text-sm">
              <h2 class="font-medium">Dimensions</h2>
              <TabsSlice dimension="uiDimension2" />
            </div>
            <div class="flex flex-wrap items-center gap-x-3 text-sm">
              <label class="font-medium">Filter by</label>
              <SelectSliceBy dimension="uiDimension3" />
              <SelectSliceBy
                dimension="uiFilter3"
                list={listFor("uiDimension3")}
                class="grow"
              />
              <button
                class="hover:underline underline-offset-4 me-4"
                onClick={() => {
                  setState("cubeActivity", "uiFilter3", "");
                }}
              >
                clear
              </button>
            </div>
          </section>
          <Dimension2 />
        </section>
        <section class="flex flex-col gap-4 ps-4 border-s border-neutral-300 dark:border-neutral-700">
          <div class="flex flex-wrap gap-x-3 text-sm">
            <label class="font-medium">Slice/Color by</label>
            <SelectSliceBy dimension="uiDimension1" />
          </div>
          <Dimension1 />
        </section>
      </section>
      <div>
        <span>URL</span> /api/v1/activity?database_id={state.database.name}
        &start=
        {state.interval_request_ms}&step=
        {Math.floor(state.interval_ms / 1000)}s&legend=
        <span class="text-green-500">{state.cubeActivity.uiDimension1}</span>
        &dim=
        <span class="text-green-500">{state.cubeActivity.uiDimension2}</span>
        &filterdim=
        <span class="text-green-500">{state.cubeActivity.uiDimension3}</span>
        &filterdimselected=
        <span class="text-green-500">
          {encodeURIComponent(state.cubeActivity.uiFilter3)}
        </span>
      </div>
      <details>
        <summary>debug</summary>

        <pre class="text-xs whitespace-pre-wrap break-works text-neutral-500 max-w-28 dark:text-neutral-400">
          Cube data
          <br />
          {JSON.stringify(state.cubeActivity, null, 2)}
        </pre>
      </details>
    </>
  );
}

function Dimension1() {
  const { state, setState } = contextState();
  return (
    <Switch fallback={<span>TODO</span>}>
      <Match when={state.cubeActivity.uiDimension1 === "waits"}>
        <section class="flex flex-col gap-4">
          <section class="flex flex-col gap-2">
            <For each={listWaits}>
              {(value, index) => {
                const isChecked =
                  state.cubeActivity.uiCheckedDimension1.includes(value);
                return (
                  <label class="flex gap-x-2 items-center">
                    <input
                      type="checkbox"
                      checked={isChecked}
                      class="aappearance-none border border-neutral-300 dark:border-neutral-700 rounded size-3.5"
                      classList={{
                        [listWaitsColors[index()]]: isChecked,
                        // "accent-red-500": !isChecked,
                      }}
                      onChange={(event) => {
                        if (event.target.checked) {
                          setState("cubeActivity", "uiCheckedDimension1", [
                            ...state.cubeActivity.uiCheckedDimension1,
                            value,
                          ]);
                        } else {
                          setState(
                            "cubeActivity",
                            "uiCheckedDimension1",
                            () => {
                              const newList =
                                state.cubeActivity.uiCheckedDimension1.filter(
                                  (v) => v !== value,
                                );
                              return newList.length ? newList : listWaits;
                            },
                          );
                        }
                      }}
                    />
                    <span class={listWaitsColors[index()]}>{value}</span>
                  </label>
                );
              }}
            </For>
          </section>
        </section>
      </Match>
      <Match when={state.cubeActivity.uiDimension1 === "sql"}>
        <div>TODO: SQL</div>
      </Match>
    </Switch>
  );
}

function Dimension2(props: { class?: string }) {
  const { state } = contextState();
  return (
    <>
      <Show
        when={state.cubeActivity.uiDimension2 === "time"}
        fallback={<DimensionView />}
      >
        <div class={`${props.class}`}>
          <DimensionTime />
        </div>
      </Show>
    </>
  );
}

// function Dimension3(props: { class?: string }) {
//   const { state } = contextState();
//
//   return (
//     <>
//       <Switch>
//         <Match when={state.cubeActivity.uiDimension3 === "time"}>
//           <div class={`${props.class}`}>
//             <DimensionTime />
//           </div>
//         </Match>
//         <Match when={state.cubeActivity.uiDimension3 === "sql"}>
//           <div>SQL</div>
//         </Match>
//         <Match when={state.cubeActivity.uiDimension3 === "waits"}>
//           <div>Waits</div>
//         </Match>
//       </Switch>
//     </>
//   );
// }

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
    // There should not be negative values in rawData
    // const rawData: State["data"]["echart1"] = state.data.echart1;
    // const totalData: number[] = [];
    // for (let i = 0; i < rawData[0].length; ++i) {
    //   let sum = 0;
    //   for (let j = 0; j < rawData.length; ++j) {
    //     sum += rawData[j][i];
    //   }
    //   totalData.push(sum);
    // }
    // const series = [
    //   "CPU",
    //   "Client:ClientRead",
    //   "Lock:tuple",
    //   "LWLock:WALWrite",
    //   "Lock:transactionid",
    // ].map((name, sid) => {
    //   return {
    //     name,
    //     type: "bar",
    //     stack: "total",
    //     barWidth: "60%",
    //     label: {
    //       show: true,
    //       formatter: (params: { value: number }) => Math.round(params.value),
    //       // Math.round(params.value * 1000) / 10 + "%",
    //     },
    //     data: rawData[sid].map(
    //       (d, did) => (totalData[did] <= 0 ? 0 : d)
    //       // totalData[did] <= 0 ? 0 : d / totalData[did]
    //     ),
    //   };
    // });
    // series.push({
    //   name: "vCPUs",
    //   type: "line",
    //   data: [20, 20, 20, 20, 20],
    //   // @ts-expect-error markLine is not defined for some reason
    //   markLine: {
    //     data: [{ type: "average", name: "Avg" }],
    //   },
    // });

    let series = [
      {
        name: "CPU",
        type: "bar",
        stack: "total",
        barWidth: "60%",
        label: {
          show: true,
          formatter: (params: { value: number }) => Math.round(params.value),
          // Math.round(params.value * 1000) / 10 + "%",
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
        // boundaryGap: [0, "100%"],
        // data: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
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

function DimensionView() {
  const { state } = contextState();
  let [random, setRandom] = createSignal(() => Math.random() * 10);

  createEffect(() => {
    state.cubeActivity.uiDimension2;
    setRandom(() => () => Math.floor(Math.random() * 10));
  });

  return (
    <section class="flex flex-col gap-4">
      <For
        each={[
          {
            len: 44.3,
            txt: "Text from query will be shown here. Colored sections will be added.",
          },
          {
            len: 24.3,
            txt: "Text from query will be shown here. Colored sections will be added.",
          },
          {
            len: 14.3,
            txt: "Text from query will be shown here. Colored sections will be added.",
          },
          {
            len: 4.3,
            txt: "Text from query will be shown here. Colored sections will be added.",
          },
          {
            len: 0.3,
            txt: "Text from query will be shown here. Colored sections will be added.",
          },
        ]}
      >
        {({ len, txt }) => <DimensionRow len={random()() + len} txt={txt} />}
      </For>
    </section>
  );
}

function DimensionRow(props: {
  len: number;
  txt:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}) {
  return (
    <section class="flex items-center">
      <div class="w-8">
        <button class="size-4 rounded-full border border-neutral-700 dark:border-neutral-300 dark:bg-black"></button>
      </div>
      <div class="w-48 xs:w-64">
        <div style={{ width: `${props.len}%` }} class="bg-red-500 rounded">
          {props.len.toFixed(1)}
        </div>
      </div>
      <div class="grow">{props.txt}</div>
    </section>
  );
}

function TabsSlice(props: { dimension: "uiDimension2" }) {
  const { state } = contextState();
  return (
    <section class="flex flex-wrap gap-3">
      <Tab
        value="time"
        txt="Time"
        selected={state.cubeActivity[props.dimension] === "time"}
      />
      <For each={listDefault}>
        {(value) => (
          <Tab
            value={value[0]}
            txt={`Top ${value[1]}`}
            selected={state.cubeActivity[props.dimension] === value[0]}
          />
        )}
      </For>
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
      onClick={() => setState("cubeActivity", "uiDimension2", props.value)}
    >
      {props.txt}
    </button>
  );
}

function SelectSliceBy(props: {
  dimension: "uiDimension1" | "uiDimension2" | "uiDimension3" | "uiFilter3";
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
      // class={`bg-transparent rounded border border-neutral-200 dark:border-neutral-700 text-fuchsia-500 ps-2 pe-8 py-1.5 hover:border-gray-400 focus:outline-none ${props.class}`}
    >
      <For each={props.list || listDefault}>
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

function listFor(dimension: string) {
  const { state } = contextState();
  if (dimension !== "uiDimension3") return [];
  let list: string[][] = [];
  switch (state.cubeActivity[dimension]) {
    case "waits":
      list = listWaits.map((x) => [x, x]);
      break;
    case "sql":
      list = ["select 1", "select 2"].map((x) => [x, x]);
      break;
    case "hosts":
      list = ["host 1", "host 2"].map((x) => [x, x]);
      break;
    case "users":
      list = ["user 1", "user 2"].map((x) => [x, x]);
      break;
    case "session_types":
      list = ["session_types 1", "session_types 2"].map((x) => [x, x]);
      break;
    case "applications":
      list = ["application 1", "application 2"].map((x) => [x, x]);
      break;
    case "databases":
      list = ["database 1", "database 2"].map((x) => [x, x]);
      break;
  }
  list.unshift(["", "no filter"]);
  return list;
}
