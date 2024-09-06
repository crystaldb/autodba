import { contextState } from "../context_state";
import { createMemo, For, Match, Switch } from "solid-js";
import {
  DimensionField,
  listColors,
  DimensionName,
  listDimensionTabNames,
  CubeData,
} from "../state";
import {
  arrange,
  distinct,
  filter,
  fixedOrder,
  map,
  tidy,
  slice,
} from "@tidyjs/tidy";
import { CubeDimensionTime } from "./cube_activity_time";
import { DimensionBars } from "./cube_activity_bars";

const cssThingy =
  "border border-zinc-200 bg-zinc-100 dark:border-zinc-600 dark:bg-zinc-800 dark:hover:bg-zinc-700 hover:bg-zinc-300 first:rounded-s-lg last:rounded-e-lg";

export type ILegend = {
  item: string;
  colorText: string;
  colorBg: string;
}[];

export function CubeActivity() {
  const { state } = contextState();

  const legendDistinct = createMemo((): ILegend => {
    state.cubeActivity.uiLegend;
    return tidy(
      state.cubeActivity.cubeData,
      distinct((row) => row.metric[state.cubeActivity.uiLegend]),
      map((row) => ({
        item: row.metric[state.cubeActivity.uiLegend],
      })),
      filter(({ item }) => !!item),
      arrange(["item"]),
      arrange([
        // move CPU to the end of the list iff it exists
        fixedOrder((row) => row.item, ["CPU"], { position: "end" }),
      ]),
      slice(0, 15),
      map((item, index) => ({
        item: item.item!,
        colorText: listColors[index]?.text || "",
        colorBg: listColors[index]?.bg || "",
      })),
    );
  });

  const cssSectionHeading = "flex flex-col gap-y-3.5";

  return (
    <section class="flex flex-col-reverse md:flex-row items-start gap-4">
      <section class={cssSectionHeading}>
        <h2 class="font-medium text-lg">Legend</h2>
        <div
          class={`flex text-sm px-2.5 py-2 border-s rounded-lg ${cssThingy}`}
        >
          <label class="whitespace-pre">Slice By:</label>
          <SelectSliceBy dimension={DimensionField.uiLegend} />
        </div>
        <Legend legend={legendDistinct()} />
      </section>

      <section class="flex flex-col gap-5">
        <section class={cssSectionHeading}>
          <h2 class="font-medium text-lg">Dimensions</h2>
          <div class="flex items-baseline gap-3 text-sm">
            <DimensionTabs
              dimension="uiDimension1"
              cubeData={state.cubeActivity.cubeData}
            />
          </div>
        </section>
        <Dimension1
          cubeData={state.cubeActivity.cubeData}
          legend={legendDistinct()}
        />
      </section>
    </section>
  );
}

interface PropsLegend {
  legend: ILegend;
}

function Legend(props: PropsLegend) {
  return (
    <section class="flex flex-col gap-4">
      <For each={props.legend}>
        {(item) => (
          <div class="flex items-center gap-x-3">
            <div class={`rounded-md size-4 ${item.colorBg}`} />
            <span class={item.colorText}>{item.item}</span>
          </div>
        )}
      </For>
    </section>
  );
}

interface IDimension1 {
  cubeData: CubeData;
  class?: string;
  legend: ILegend;
}

function Dimension1(props: IDimension1) {
  const { state } = contextState();
  return (
    <>
      <Switch>
        <Match when={state.cubeActivity.uiDimension1 === DimensionName.time}>
          <div class={`${props.class}`}>
            <CubeDimensionTime legend={props.legend} />
          </div>
        </Match>
        <Match when={true}>
          <DimensionBars
            cubeData={props.cubeData}
            legend={props.legend}
            // class="min-h-64"
          />
        </Match>
      </Switch>
    </>
  );
}

interface IDimensionTabs {
  dimension: "uiDimension1";
  cubeData: CubeData;
}

function DimensionTabs(props: IDimensionTabs) {
  const { state } = contextState();

  return (
    <section class="flex flex-col gap-3">
      <section class="flex flex-wrap gap-y-2">
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
      </section>
      {/*
      <section>
        <section class="flex gap-x-3 text-sm">
          <label class="font-medium me-6">Filter by</label>
          <SelectSliceBy
            dimension={DimensionField.uiFilter1}
            list={[["none", "No filter"], ...listDimensionTabNames()]}
          />
          <Show when={state.cubeActivity.uiFilter1 !== DimensionName.none}>
            <div class="self-end flex flex-wrap items-center gap-x-3 text-sm">
              <SelectSliceBy
                dimension="uiFilter1Value"
                list={listFor(DimensionField.uiFilter1, props.cubeData)}
                class="grow max-w-screen-sm"
              />
              <button
                class="hover:underline underline-offset-4 me-4"
                onClick={() => {
                  setState("cubeActivity", "uiFilter1Value", "");
                }}
              >
                clear
              </button>
            </div>
          </Show>
        </section>
      </section>
      */}
    </section>
  );
}

function Tab(props: { value: string; txt: string; selected: boolean }) {
  const { setState } = contextState();
  return (
    <button
      value={props.value}
      class={`tracking-wider flex text-sm px-6 py-2 font-normala ${cssThingy}`}
      classList={{
        "font-semibold text-fuchsia-500 bg-zinc-300 dark:bg-zinc-700":
          props.selected,
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
      class={`bg-transparent text-fuchsia-500 px-2 focus:outline-none ${props.class}`}
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
