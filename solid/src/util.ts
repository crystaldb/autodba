export function unique<T>(list: T[]): T[] {
  return [...new Set(list)];
}

export function truncateString(str: string, maxLength = 100) {
    if (str.length > maxLength) {
        return str.substring(0, maxLength - 3) + '...';
    }
    return str;
}

// function spliceArraysTogetherSkippingDuplicateTimestamps(
//   arr1: any[],
//   arr2: any[] = [],
// ): any[] {
//   // in the `arr1` array, starting at the end of the array and looking back up to `arr2.length` items, remove any timestamps from `arr1` that are already present in `arr2`, and then append the new `arr2` array to the end of `arr1`.
//   if (arr1.length === 0) return arr2;
//   if (arr2.length === 0) return arr1;
//   const newTimestamps = new Set(
//     arr2.map((row: { time_ms: any }) => row.time_ms),
//   );
//   const rangeStart = Math.max(0, arr1.length - arr2.length);
//
//   let insertAt = arr1.length;
//   for (let i = arr1.length - 1; i >= rangeStart; --i) {
//     if (newTimestamps.has(arr1[i].time_ms)) {
//       insertAt = i;
//     }
//   }
//   return [...arr1.slice(0, insertAt), ...arr2];
// }
