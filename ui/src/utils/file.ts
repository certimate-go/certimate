export const readFileAsText = (file: File): Promise<string> => {
  const { promise, resolve, reject } = Promise.withResolvers<string>();

  const reader = new FileReader();
  reader.onload = () => {
    if (reader.result != null) {
      resolve(reader.result.toString());
    } else {
      reject(new Error("Read file failed: result is null"));
    }
  };
  reader.onerror = () => reject(reader.error);
  reader.readAsText(file, "utf-8");

  return promise;
};
