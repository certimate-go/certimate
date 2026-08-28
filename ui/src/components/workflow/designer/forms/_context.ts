import { createContext, useContext } from "react";
import { type FlowNodeEntity } from "@flowgram.ai/fixed-layout-editor";

export {
  FormNestedFieldsContext,
  FormNestedFieldsContextProvider,
  useFormNestedFieldsContext,
  type FormNestedFieldsContextType,
} from "@/components/_shared/formNestedFieldsContext";

// #region NodeFormContext
export type NodeFormContextType = {
  node: FlowNodeEntity;
};

export const NodeFormContext = createContext<NodeFormContextType>({} as NodeFormContextType);

export const NodeFormContextProvider = NodeFormContext.Provider;

export const useNodeFormContext = () => {
  const context = useContext(NodeFormContext);
  if (!context) {
    throw new Error("`NodeFormContext` must be used within a `NodeFormContextProvider`");
  }
  return context;
};
// #endregion
