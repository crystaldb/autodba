export function unique<T>(list: T[]): T[] {
  return [...new Set(list)];
}

export function truncateString(str: string, maxLength = 100) {
  if (str.length > maxLength) {
    return `${str.substring(0, maxLength - 3)}...`;
  }
  return str;
}
