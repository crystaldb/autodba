import { contextState } from "../context_state";
import { createMemo, createResource, For, JSX } from "solid-js";
import {
  DimensionField,
  DimensionName,
  listDimensionTabNames,
  CubeData,
} from "../state";
import { first, groupBy, sum, summarize, tidy } from "@tidyjs/tidy";
import { ILegend } from "./cube_activity";
import { queryCube } from "../http";

interface IDimensionBars {
  cubeData: CubeData;
  legend: ILegend;
  class?: string;
}

export function DimensionBars(props: IDimensionBars) {
  const { state, setState } = contextState();
  const changed = createMemo((changeCount: number) => {
    state.range_begin;
    state.range_end;
    state.database_instance.dbidentifier;
    state.cubeActivity.uiLegend;
    state.cubeActivity.uiDimension1;
    state.cubeActivity.uiFilter1;
    state.cubeActivity.uiFilter1Value;
    console.log("changed", changeCount);
    return changeCount + 1;
  }, 0);

  createResource(changed, () => {
    queryCube(state, setState);
  });

  const cubeDataGrouped = createMemo(
    (): {
      dimensionValue: string;
      total: number;
      records: {
        metric: { [key: string]: string };
        values: { value: any }[];
      }[];
    }[] => {
      if (state.cubeActivity.uiDimension1 === DimensionName.time) {
        return [];
      }
      let cubeData = tidy(
        props.cubeData,
        // filter( (d) => !!d.metric[state.cubeActivity.uiDimension1] && (d.metric[state.cubeActivity.uiDimension1] === DimensionName.time || !!d.metric[state.cubeActivity.uiDimension1]),),
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
      // console.log("cubeDataGrouped", cubeData.length);
      return cubeData;
    },
  );

  return (
    <section class={`flex flex-col gap-4 ${props.class}`}>
      <For each={cubeDataGrouped()}>
        {({ total, dimensionValue, records }) => (
          <DimensionRowGrouped
            len={total}
            txt={dimensionValue}
            records={records}
            legend={props.legend}
          />
        )}
      </For>
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

function DimensionRowGrouped(props: IDimensionRow) {
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
  cubeData: CubeData;
}

function DimensionTabs(props: IDimensionTabs) {
  const { state, setState } = contextState();

  return (
    <section class="flex flex-col gap-3">
      <section class="flex gap-3 justify-between">
        <h2 class="font-medium">Dimensions</h2>
        <div class="flex flex-wrap gap-3">
          <Tab
            value={DimensionName.time}
            txt="Time"
            selected={
              state.cubeActivity[props.dimension] === DimensionName.time
            }
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
        </div>
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

// function listFor(
//   dimensionField: DimensionField,
//   cubeData: () => CubeData,
// ): [string, string][] {
//   if (dimensionField !== DimensionField.uiFilter1) return [];
//   const { state } = contextState();
//   const dimensionName: DimensionName = state.cubeActivity[dimensionField];
//
//   const input = tidy(
//     cubeData(),
//     filter((d) => !!d.metric[dimensionName]),
//     map((d) => ({ result: d.metric[dimensionName] })),
//     distinct((d) => d.result),
//     filter(({ result }) => !!result),
//   ).map(({ result }) => result!);
//
//   let list: [string, string][] = input.map((x) => [x, x]);
//
//   list.unshift(["", "no filter"]);
//   return list;
// }
