import { createFakeCubeActivityArrays } from "./fake.cubeactivity";
import { ArraysCubeActivity } from "./http";

let timeB = 0;
let dateTimeDay = "2009/10/18";
let dateTimeHour = 8;
let baseDate = +new Date(1988, 9, 3);
let oneDay = 24 * 3600 * 1000;

export function httpFakeDatabase(): {
  database: {
    name: string;
    engine: string;
    version: string;
    size: string;
    kind: string;
  };
} {
  return {
    database: {
      name: "mohammad-dashti-rds-1",
      engine: "PostgreSQL",
      version: "16.3",
      size: "db.t4g.medium",
      kind: "Instance",
    },
  };
}

export function httpFakeCubeActivity(): ArraysCubeActivity {
  if (timeB === 0) {
    timeB = Date.now();
    return createFakeCubeActivityArrays(1);
  } else {
    return createFakeCubeActivityArrays(2);
  }
}

export function httpFake(): {
  echart1: number[][];
  echart2a: string[];
  echart2b: number[];
  echart2c: number[];
  cpu: number[];
  time: number[];
} {
  if (timeB === 0) {
    timeB = Date.now();
    return {
      cpu: [10, 20, 30, 50],
      time: [
        +new Date((baseDate += oneDay)),
        +new Date((baseDate += oneDay)),
        +new Date((baseDate += oneDay)),
        +new Date((baseDate += oneDay)),
      ],

      echart1: [
        [10, 30, 30, 33, 30, 30, 30],
        [32, 13, 10, 13, 9, 23, 21],
        [22, 18, 19, 23, 20, 30, 30],
        [15, 21, 20, 15, 10, 30, 40],
        [82, 83, 90, 93, 90, 30, 20],
      ],
      echart2a: dataA().slice(0, 4),
      echart2b: dataB().slice(0, 4),
      echart2c: dataC().slice(0, 4),
    };
  } else {
    timeB = Date.now();
    let dataA = [];
    let a = dateTimeDay + "\n" + ++dateTimeHour + ":00";
    dataA.push(a);
    let dataB = [];
    let b = Math.random() * 10;
    dataB.push(b);
    let dataC = [];
    let c = Math.random() * 100;
    dataC.push(c);

    return {
      cpu: [Math.floor(c)],
      time: [+new Date((baseDate += oneDay))],
      echart1: [
        [tweakValueAt()],
        [tweakValueAt()],
        [tweakValueAt()],
        [tweakValueAt()],
        [tweakValueAt()],
      ],
      echart2a: dataA,
      echart2b: dataB,
      echart2c: dataC,
    };
  }
}

function tweakValueAt() {
  const newVal = Math.floor(Math.random() * 100 + 0.5);
  return Math.max(0, newVal);
}

function dataA() {
  // prettier-ignore
  return [
    "2009/6/12 2:00",
    "2009/6/12 3:00",
  ].map(function (str) {
    return str.replace(" ", "\n");
  });
}

function dataB() {
  // prettier-ignore
  return [
    0.97, 0.96,
  ];
}

function dataC() {
  // prettier-ignore
  return [
    0, 0,
  ];
}
