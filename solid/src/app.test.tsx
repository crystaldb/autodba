import { test, expect } from "vitest";
import { render } from "@solidjs/testing-library";
import App from "./App";
import { test_setup } from "./test.setup";

test_setup();

test("it will render", async () => {
  const { getByText } = render(() => <App />);
  expect(
    getByText("[Crystal Observability homepage stuff here]"),
  ).toBeInTheDocument();
});
