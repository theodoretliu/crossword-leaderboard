export const secondsToMinutes = (seconds: number) => {
  const minutes = Math.floor(seconds / 60);
  const secondsRemaining = seconds % 60;

  return `${minutes}:${secondsRemaining.toString().padStart(2, "0")}`;
};

export function padRight<T>(arr: Array<T>, value: T, length: number): Array<T> {
  let newArr = arr.slice();

  while (newArr.length < length) {
    newArr.push(value);
  }

  return newArr;
}
