import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconCloudUpload } from "@tabler/icons-react";

import { newNode } from "@/domain/workflow";
import { getCertificateSubjectAltNames as getX509SubjectAltNames } from "@/utils/x509";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BizUploadNodeConfigForm from "../forms/BizUploadNodeConfigForm";

export const BizUploadNodeRegistry: NodeRegistry = {
  type: NodeType.BizUpload,

  kind: NodeKindType.Business,

  meta: {
    labelText: getI18n().t("workflow_node.upload.label"),

    icon: IconCloudUpload,
    iconColor: "#fff",
    iconBgColor: "#5b65f5",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BizUploadNodeConfigForm.getSchema({}).safeParse(value);
        if (!res.success) {
          return {
            message: res.error.message,
            level: FeedbackLevel.Error,
          };
        }
      },
    },

    render: () => {
      const { t } = getI18n();

      return (
        <BaseNode
          description={
            <Field<string> name="config.source">
              {({ field: { value: fieldSource } }) => (
                <>
                  {fieldSource == null || fieldSource === "" || fieldSource === "form" ? (
                    <Field<string> name="config.certificate">
                      {({ field: { value: fieldCertificate } }) => {
                        const displayText = fieldCertificate ? getX509SubjectAltNames(fieldCertificate).join(";") : void 0;
                        return <>{displayText || t("workflow.detail.design.editor.placeholder")}</>;
                      }}
                    </Field>
                  ) : (
                    <Field<string> name="config.certificate">
                      {({ field: { value: fieldCertificate } }) => <>{fieldCertificate || t("workflow.detail.design.editor.placeholder")}</>}
                    </Field>
                  )}
                </>
              )}
            </Field>
          }
        />
      );
    },
  },

  onAdd: () => {
    return newNode(NodeType.BizUpload, { i18n: getI18n() });
  },
};
