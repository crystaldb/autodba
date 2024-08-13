/** @import {JSX} from 'solid-js' */

// import { contextState } from "../context_state";

import { A } from "@solidjs/router";
import { JSX } from "solid-js";

function NavTop(props: {
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
        class={`flex flex-col xs:flex-row items-center justify-between h-16 pe-3 ${props.class}`}
      >
        <div class="flex items-center">
          <A href="/" class="dark:hidden flex items-center" end>
            <img src="/logo-dark-text.svg" alt="logo" class="h-12" />
            <span class="text-lg font-medium">AutoDBA</span>
          </A>
          <A href="/" class="hidden dark:flex items-center" activeClass="">
            <img src="/logo-light-text.svg" alt="logo" class="h-12" />
            <span class="text-lg font-medium">AutoDBA</span>
          </A>
        </div>
        <section class="flex gap-x-4 items-center">{props.children}</section>
      </nav>
    </>
  );
}

export default NavTop;
