export function unique<T>(list: T[]): T[] {
  return [...new Set(list)];
}

export function truncateString(str, maxLength = 100) {
    console.log('Truncate string');
    console.log('str', str);
    if (str.length > maxLength) {
        return str.substring(0, maxLength - 3) + '...';
    }
    return str;
}
