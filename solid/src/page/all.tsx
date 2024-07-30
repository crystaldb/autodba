/** @import {JSX} from "solid-js" */
import { batch, createEffect, Show } from "solid-js";
import { contextState } from "../context_state";
import { ViewTitle } from "../view/common";
import { ViewTextinput } from "../view/input_text";

/**
 * @param {Object} props
 * @returns {JSX.Element}
 */

export function PageAll(props: any) {
  const { state, setState } = contextState();

  return (
    <>
      <ViewTitle title="Charts of all kinds" />
      <ViewTextinput title="All Charts" />
    </>
  );
}
