/** @import {JSX} from 'solid-js' */

// import { contextState } from "../context_state";

import { A } from "@solidjs/router";
import { JSX } from "solid-js";

export function NavTop(props: {
  class?: string;
  children:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}) {
  // const { state } = contextState();
  return (
    <>
      <nav
        class={`flex flex-col lg:flex-row items-center justify-between h-16 pe-3 border-b border-zinc-500 ${props.class}`}
      >
        <A href="/" class="flex items-center gap-2" end>
          <img src="/logo.svg" alt="logo" class="h-7" />
          <span class="text-2xl font-medium">AutoDBA</span>
        </A>
        <section class="flex flex-wrap gap-4 items-center">
          {props.children}
        </section>
      </nav>
    </>
  );
}
