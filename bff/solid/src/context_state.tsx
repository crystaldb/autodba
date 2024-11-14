import { createContext, useContext } from "solid-js";
import type { SetStoreFunction } from "solid-js/store";
import { type Anything, type State, useState } from "~/state";

const StateContext = createContext(useState());

export function ContextState(props: { children: Anything }) {
  return (
    <StateContext.Provider value={useState()}>
      {props.children}
    </StateContext.Provider>
  );
}

export function contextState(): {
  state: State;
  setState: SetStoreFunction<State>;
} {
  return useContext(StateContext);
}
