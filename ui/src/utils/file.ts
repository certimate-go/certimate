export function readFileAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = () => {
      if (reader.result) {
        resolve(reader.result.toString());
      } else {
        reject("No content found");
      }
    };

    reader.onerror = () => reject(reader.error);

    reader.readAsText(file, "utf-8");
  });
}
