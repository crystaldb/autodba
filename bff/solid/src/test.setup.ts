// import { vi } from "vitest";

export function test_setup() {
  // REMOVING FOR NOW UNTIL WE START TESTING WITH THIS AGAIN: 2024.12.03
  // const ResizeObserverMock = vi.fn(() => ({
  //   observe: vi.fn(),
  //   unobserve: vi.fn(),
  //   disconnect: vi.fn(),
  // }));
  // vi.stubGlobal("ResizeObserver", ResizeObserverMock);
  //
  // // HTMLCanvasElement.prototype.getContext = vi.fn();
  // // @ts-expect-error
  // window.HTMLCanvasElement.prototype.getContext = () => ({
  //   arc() {},
  //   beginPath() {},
  //   bezierCurveTo() {},
  //   clearRect() {},
  //   clip() {},
  //   closePath() {},
  //   createImageData() {
  //     return [];
  //   },
  //   drawImage() {},
  //   fill() {},
  //   fillRect() {},
  //   fillText() {},
  //   getImageData(x, y, w, h) {
  //     return { data: new Array(w * h * 4) };
  //   },
  //   lineTo() {},
  //   measureText() {
  //     return { width: 0 };
  //   },
  //   moveTo() {},
  //   putImageData() {},
  //   rect() {},
  //   restore() {},
  //   rotate() {},
  //   save() {},
  //   scale() {},
  //   setTransform() {},
  //   stroke() {},
  //   transform() {},
  //   translate() {},
  // });
  //
  // window.HTMLCanvasElement.prototype.toDataURL = () => "";
}
