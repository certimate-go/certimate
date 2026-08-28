import { createContext, useContext } from "react";

export type FormNestedFieldsContextType = {
  parentNamePath: string;
};

export const FormNestedFieldsContext = createContext<FormNestedFieldsContextType>({
  parentNamePath: "",
});

export const FormNestedFieldsContextProvider = FormNestedFieldsContext.Provider;

export const useFormNestedFieldsContext = () => {
  const context = useContext(FormNestedFieldsContext);
  if (!context) {
    throw new Error("`FormNestedFieldsContext` must be used within a `FormNestedFieldsContextProvider`");
  }
  return context;
};
