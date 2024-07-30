export async function getData(
  setState: (arg0: string, arg1: any) => void
): Promise<string> {
  const response = await fetch("/api/data", {
    method: "GET",
  });
  return response.json().then((json) => {
    setState("data", json.data);
    return json.email;
  });
}
