import { SolidMarkdown } from "solid-markdown";
import { HiOutlineEyeSlash, HiOutlineQuestionMarkCircle } from "solid-icons/hi";
import { JSX, Show } from "solid-js";

export function ViewTitle(props: {
  title:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}) {
  return (
    <span class="font-medium dark:font-base sm:font-base sm:text-sm/6 leading-6">
      {props.title}
    </span>
  );
}
