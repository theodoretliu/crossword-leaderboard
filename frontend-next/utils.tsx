export const secondsToMinutes = (seconds: number) => {
  const minutes = Math.floor(seconds / 60);
  const secondsRemaining = seconds % 60;

  return `${minutes}:${secondsRemaining.toString().padStart(2, "0")}`;
};
