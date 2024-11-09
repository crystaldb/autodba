import { contextState } from "../context_state";
import { createMemo, For, JSX } from "solid-js";
import { DimensionName, CubeData, ApiEndpoint, LegendData } from "../state";
import { first, groupBy, sum, summarize, tidy } from "@tidyjs/tidy";
import { ILegend } from "./cube_activity";

interface IDimensionBars {
  cubeData: CubeData;
  legend: ILegend;
  class?: string;
}

export function DimensionBars(props: IDimensionBars) {
  const { state, setState } = contextState();
  setState("apiThrottle", "needDataFor", ApiEndpoint.activity);

  const cubeDataGrouped = createMemo<
    {
      dimensionValue: string;
      total: number;
      records: {
        metric: { [key: string]: string };
        values: { value: number }[];
      }[];
    }[]
  >(() => {
    if (state.activityCube.uiDimension1 === DimensionName.time) {
      return [];
    }
    const cubeData = tidy(
      props.cubeData,
      groupBy(
        (d) => d.metric[state.activityCube.uiDimension1],
        [
          summarize({
            dimensionValue: first(
              (d: {
                metric: Record<string, string>;
                values: { value: number }[];
              }) => {
                return d.metric[state.activityCube.uiDimension1];
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
    return cubeData;
  });

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
  records: LegendData;
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
              txt={
                record.metric[state.activityCube.uiLegend] || "unknown-metric"
              }
              legend={props.legend}
            />
          )}
        </For>
        <p class="ms-2 me-3">
          {props.records
            .reduce(
              (sum: number, record: { values: { value: number }[] }) =>
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
  txt: string;
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
