/** @import {JSX} from 'solid-js' */

// import { contextState } from "../context_state";

import { A } from "@solidjs/router";
import { JSX } from "solid-js";

function NavTop(props: {
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
  // <img src="/logo.webp" alt="logo" class="bg-white w-[652px] h-[135px]" />
  return (
    <>
      <nav class="flex items-center justify-between h-16 px-2">
        <A href="/">
          <img
            src="/logo.webp"
            alt="logo"
            class="bg-white w-[163px] h-[34px]"
          />
        </A>
        <section class="flex gap-x-4 items-center">{props.children}</section>
      </nav>
    </>
  );
}

export default NavTop;
