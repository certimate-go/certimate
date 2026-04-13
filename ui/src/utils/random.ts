export function randomString(): string;
export function randomString(length: number): string;
export function randomString(options: { length?: number; alphabet?: string }): string;
export function randomString(arg?: number | { length?: number; alphabet?: string }): string {
  const DEFAULT_LENGTH = 32;
  const DEFAULT_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";

  let length = DEFAULT_LENGTH;
  let alphabet = DEFAULT_ALPHABET;

  if (typeof arg === "number") {
    length = arg;
  } else if (arg && typeof arg === "object") {
    if (arg.length != null) {
      length = arg.length;
    }
    if (arg.alphabet != null) {
      alphabet = arg.alphabet;
    }
  }

  const randomBytes = crypto.getRandomValues(new Uint8Array(length));
  return Array.from(randomBytes, (b) => alphabet[b % alphabet.length]).join("");
}
