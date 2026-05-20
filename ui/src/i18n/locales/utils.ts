const flatten = (raw: Record<string, any>): Record<string, string> => {
  const flatten = (obj: Record<string, any>, prefix = ""): Record<string, string> => {
    return Object.keys(obj)
      .filter((prop) => !prop.startsWith("$"))
      .reduce(
        (acc, prop) => {
          const key = (prefix ? `${prefix}${prop}` : prop).replace(/\.$/, "");
          const value = obj[prop];
          if (typeof value === "object" && value != null) {
            Object.assign(acc, flatten(value, `${key}.`));
          } else {
            if (acc[key]) {
              console.warn(`[certimate] i18n: duplicate translation key "${key}" with value "${value}" is overwritten by previous value "${acc[key]}"`);
            }
            acc[key] = value;
          }
          return acc;
        },
        {} as Record<string, string>
      );
  };

  const ns = raw["$ns"] ? `${raw["$ns"]}.` : "";
  return flatten(raw, ns);
};

const merge = (...translations: Record<string, string>[]): Record<string, string> => {
  return translations.reduce(
    (acc, translation) => {
      for (const key in translation) {
        if (acc[key] && acc[key] !== translation[key]) {
          console.warn(`[certimate] i18n: duplicate translation key "${key}" with value "${translation[key]}" is overwritten by previous value "${acc[key]}"`);
        }
        acc[key] = translation[key];
      }

      return acc;
    },
    {} as Record<string, string>
  );
};

export const buildTranslations = (...raws: Record<string, any>[]): Record<string, string> => {
  const translations = raws.map(flatten);
  return merge(...translations);
};
