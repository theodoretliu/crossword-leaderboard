import dayjs from "dayjs";

export const secondsToMinutes = (
  seconds: number,
  showDecimals: boolean = false
) => {
  let secondsStr: string;
  if (showDecimals) {
    secondsStr = seconds.toFixed(2);
    const decimals = secondsStr.slice(-3);
    const totalSeconds = parseInt(secondsStr.slice(0, -3));
    const minutes = Math.floor(totalSeconds / 60);
    const secondsRemaining = totalSeconds % 60;

    return `${minutes}:${secondsRemaining
      .toString()
      .padStart(2, "0")}${decimals}`;
  }

  secondsStr = seconds.toFixed(0);
  const totalSeconds = parseInt(secondsStr);
  const minutes = Math.floor(totalSeconds / 60);
  const secondsRemaining = totalSeconds % 60;

  return `${minutes}:${secondsRemaining.toString().padStart(2, "0")}`;
};

export function padRight<T>(arr: Array<T>, value: T, length: number): Array<T> {
  let newArr = arr.slice();

  while (newArr.length < length) {
    newArr.push(value);
  }

  return newArr;
}

export function dateToFormat(date: string) {
  return dayjs(date).utc().format("dddd, M/D");
}

export function datesToFormat(dates: Array<string>) {
  return dates.map(dateToFormat);
}
