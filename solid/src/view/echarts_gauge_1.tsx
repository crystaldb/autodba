import { EChartsAutoSize } from "echarts-solid";
import { contextState } from "../context_state";
import { mergeProps } from "solid-js";
import { datazoom } from "../state";

interface IProps {
  data: number[];
}

export function EchartsGauge1(props: IProps) {
  let ref: import("@solid-primitives/refs").Ref<HTMLDivElement>;
  const { state, setState } = contextState();
  const base = {
    dataset: { source: [[0]] },
    backgroundColor: "transparent",
    series: [
      {
        type: "gauge",
        radius: "175%",
        center: ["50%", "88%"],
        startAngle: 180,
        endAngle: 0,
        min: 0,
        max: 60,
        splitNumber: 6,
        itemStyle: {
          // color: "#FFAB91",
          color: "fuchsia",
        },
        pointer: {
          show: false,
        },
        progress: {
          show: true,
          overlap: true,
          width: 5,
          itemStyle: {
            // color: "black",
          },
        },
        axisLine: {
          lineStyle: {
            width: 5,
            color: "#555",
          },
        },
        axisTick: {
          distance: 5,
          splitNumber: 5,
          lineStyle: {
            width: 1,
            color: "#555",
          },
        },
        splitLine: {
          distance: 5,
          length: 10,
          lineStyle: {
            width: 1,
            color: "#555",
          },
        },
        axisLabel: {
          show: true,
          distance: 10,
          color: "#555",
          fontSize: 10,
        },
        anchor: {
          show: false,
        },
        title: {
          show: false,
        },
        detail: {
          valueAnimation: false,
          // width: "60%",
          // lineHeight: 40,
          // borderRadius: 4,
          offsetCenter: [0, "-7%"],
          fontSize: 23,
          fontWeight: "bolder",
          formatter: (value: number) => (value ? `${value.toFixed(1)} %` : ""),
          color: "inherit",
        },
      },
      // {
      //   type: "gauge",
      //   center: ["50%", "50%"],
      //   startAngle: 200,
      //   endAngle: -20,
      //   min: 0,
      //   max: 60,
      //   itemStyle: {
      //     color: "#FD7347",
      //   },
      //   progress: {
      //     show: true,
      //     width: 8,
      //   },
      //   pointer: {
      //     show: false,
      //   },
      //   axisLine: {
      //     show: false,
      //     // lineStyle: {
      //     //   color: [
      //     //     [0.1, "red"],
      //     //     [0.2, "orange"],
      //     //     [0.8, "green"],
      //     //     [1, "blue"],
      //     //   ],
      //     // },
      //   },
      //   axisTick: {
      //     show: false,
      //   },
      //   splitLine: {
      //     show: false,
      //   },
      //   axisLabel: {
      //     show: false,
      //   },
      //   detail: {
      //     show: false,
      //   },
      // },
    ],
  };

  const eventHandlers = {
    click: (event: any) => {
      console.log("Chart is clicked!", event);
    },
    highlight: (event: any) => {
      console.log("Chart Highlight", event);
    },
    datazoom: datazoom.bind(null, setState, state),
  };

  return (
    <>
      <EChartsAutoSize
        // @ts-expect-error suppress complaint about `type: "gauge"`
        option={mergeProps(base, {
          dataset: {
            source: [[props.data[props.data.length - 1]]],
          },
        })}
        eventHandlers={eventHandlers}
        ref={ref}
        class=""
      />
    </>
  );
}

// option ={
//   series: [
//     {
//       data: [[30]],
//       type: 'gauge',
//       startAngle: 180,
//       endAngle: 0,
//       min: 0,
//       max: 100,
//       splitNumber: 3,
//       pointer: {
//         icon: 'circle',
//         length: '12%',
//         width: 50,
//         offsetCenter: [0, '-90%'],
//         itemStyle: {
//           color: '#FFFFFF',
//           borderColor: 'black',
//           borderWidth: 5,
//           shadowColor: 'rgba(10, 31, 68, 0.5)',
//           shadowBlur: 2,
//           shadowOffsetY: 1,
//         },
//       },
//       axisLine: {
//         show: true,
//         roundCap: false,
//         lineStyle: {
//           width: 16,
//           color: [
//             [0.1, "red"],
//             [0.11],
//             [0.2, "red"],
//             [0.21],
//             [0.5, '#e76262'],
//             [0.54],
//             [0.66, '#f9cf4a'],
//             [0.7],
//             [0.83, '#eca336'],
//             [0.87],
//             [1, '#3ece80'],
//           ],
//         },
//       },
//       axisTick: {
//         length: 2,
//         lineStyle: {
//           color: '#8a94a6',
//           width: 2,
//         },
//       },
//       splitLine: {
//         show: false,
//       },
//       axisLabel: {
//         show: false,
//       },
//       title: {
//         show: false,
//       },
//       detail: {
//         rich: {
//           header: {
//             fontSize: 36,
//             fontWeight: 700,
//             fontFamily: 'Open Sans',
//             color: '#0a1f44',
//           },
//           subHeader: {
//             fontSize: 16,
//             fontWeight: 400,
//             fontFamily: 'Open Sans',
//             color: '#8a94a6',
//           },
//         },
//         formatter: ['{header|{value}}', '{subHeader|15-30-2022}'].join('\n'),
//         offsetCenter: [0, '-20%'],
//         valueAnimation: true,
//       },
//     },
//   ],
// }
