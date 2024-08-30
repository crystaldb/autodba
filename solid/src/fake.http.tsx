export function httpFakeDatabase(): {
  database: {
    name: string;
    engine: string;
    version: string;
    size: string;
    kind: string;
    datname: string;
  };
} {
  return {
    database: {
      name: "mohammad-dashti-rds-1",
      engine: "PostgreSQL",
      version: "16.3",
      size: "db.t4g.medium",
      kind: "Instance",
      datname: "(postgres|rdsadmin|template0|template1)",
    },
  };
}
