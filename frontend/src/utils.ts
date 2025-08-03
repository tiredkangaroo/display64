export async function waitForWindowClose(popup: Window | null, interval: number = 200): Promise<void> {
  if (!popup) {
    throw new Error("Popup window is null. It may have been blocked.");
  }

  return new Promise((resolve) => {
    const pollTimer = setInterval(() => {
      if (popup.closed) {
        clearInterval(pollTimer);
        resolve();
      }
    }, interval);
  });
}