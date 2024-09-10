import { contextState } from "../context_state";
import { createMemo, createResource, For, JSX, Show } from "solid-js";
import { DimensionName, CubeData } from "../state";
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
    // state.timeframe_ms; // handled by createEffect locally
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

  const [resourceChanged] = createResource(changed, () => {
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
      <Show when={resourceChanged} keyed>
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
      </Show>
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
    <section data-testclass="dimensionRow" class="flex items-center">
      <div class="w-48 xs:w-64 flex flex-row items-center">
        <For each={props.records}>
          {(record) => (
            <DimensionRowPart
              len={record.values[0].value}
              txt={record.metric[state.cubeActivity.uiLegend]}
              legend={props.legend}
            />
          )}
        </For>
        <p class="ms-2 me-3">
          {props.records
            .reduce(
              (sum: number, record: { values: { value: any }[] }) =>
                sum + record.values[0].value,
              0,
            )
            .toFixed(1)}
        </p>
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
      style={{ width: `${props.len * 15}px` }}
      class={`flex items-center text-sm ps-0.5 py-1 cursor-default h-8 ${css}`}
      title={props.txt}
    >
      {props.len >= 2 ? props.len.toFixed(1) : ""}
    </div>
  );
}
