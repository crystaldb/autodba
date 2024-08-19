import { faker } from "@faker-js/faker";
import { ArraysCubeActivity } from "./http";
import { Waits } from "./state";

type SessionTypes = "client backend" | "autovacuum worker";

interface CubeActivity {
  valActiveSessionCount: number;
  valTime: number;
  valSql: string;
  valWaits: Waits;
  valHosts: string;
  valUsers: string;
  valSession_types: SessionTypes;
  valApplications: string;
  valDatabases: string;
}

export function createFakeCubeActivityArrays(
  count: number,
): ArraysCubeActivity {
  let arrays: ArraysCubeActivity = {
    arrActiveSessionCount: [],
    arrTime: [],
    arrSql: [],
    arrWaits: [],
    arrHosts: [],
    arrUsers: [],
    arrSession_types: [],
    arrApplications: [],
    arrDatabases: [],
  };
  new Array(count)
    .fill(undefined)
    .map(() => createFakeCubeActivity())
    .map((item) => {
      arrays.arrActiveSessionCount.push(item.valActiveSessionCount);
      arrays.arrTime.push(item.valTime);
      arrays.arrSql.push(item.valSql);
      arrays.arrWaits.push(item.valWaits);
      arrays.arrHosts.push(item.valHosts);
      arrays.arrUsers.push(item.valUsers);
      arrays.arrSession_types.push(item.valSession_types);
      arrays.arrApplications.push(item.valApplications);
      arrays.arrDatabases.push(item.valDatabases);
    });

  return arrays;
}

function createFakeCubeActivity(): CubeActivity {
  return {
    valActiveSessionCount: faker.number.int(10),
    valTime: faker.date.recent().getTime(),
    valSql: faker.lorem.sentence(),
    valWaits: faker.helpers.weightedArrayElement([
      { weight: 1, value: "LWLock:BufferContent" },
      { weight: 1, value: "LWLock:WALInsert" },
      { weight: 1, value: "Timeout:VaccumDelay" },
      { weight: 2, value: "Timeout:VaccumTruncate" },
      { weight: 3, value: "Client:ClientRead" },
      { weight: 8, value: "IO:WALSync" },
      { weight: 10, value: "Lock:tuple" },
      { weight: 20, value: "LWLock:WALWrite" },
      { weight: 80, value: "Lock:transaactionid" },
      { weight: 10, value: "CPU" },
    ]),
    valHosts: faker.helpers.weightedArrayElement([
      { weight: 90, value: "172.31.56.138" },
      { weight: 5, value: "Unknown" },
      { weight: 1, value: "73.14.92.98" },
      // faker.internet.ip(),
    ]),
    valUsers: faker.helpers.weightedArrayElement([
      { weight: 90, value: "postgres" },
      { weight: 10, value: "Unknown" },
      { weight: 3, value: "User1" },
      { weight: 2, value: "rdsadmin" },
    ]),
    valSession_types: faker.helpers.weightedArrayElement([
      { weight: 90, value: "client backend" },
      { weight: 3, value: "autovacuum worker" },
    ]),
    valApplications: faker.helpers.weightedArrayElement([
      { weight: 90, value: "app1" },
      { weight: 3, value: "Unknown" },
      { weight: 2, value: "app2" },
    ]),
    valDatabases: faker.helpers.weightedArrayElement([
      { weight: 90, value: "postgres" },
      { weight: 10, value: "db1" },
      { weight: 5, value: "rdsadmin" },
      { weight: 1, value: "template0" },
    ]),
  };
}
