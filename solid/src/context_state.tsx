import { createContext, JSX, useContext } from "solid-js";
import { useState, State } from "./state";

const StateContext = createContext(useState());

export function ContextState(props: {
  children:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}) {
  return (
    <StateContext.Provider value={useState()}>
      {props.children}
    </StateContext.Provider>
  );
}

export function contextState(): { state: State; setState: any } {
  return useContext(StateContext);
}
