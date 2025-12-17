import { getI18n, useTranslation } from "react-i18next";
import { Form, Input, Radio, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import Show from "@/components/Show";
import { validDomainName } from "@/utils/validators";

import { useFormNestedFieldsContext } from "./_context";

const DOMAIN_MATCH_PATTERN_EXACT = "exact" as const;
const DOMAIN_MATCH_PATTERN_WILDCARD = "wildcard" as const;
const DOMAIN_MATCH_PATTERN_CERTSAN = "certsan" as const;

const DOMAIN_TYPE_PLAY = "play" as const;
const DOMAIN_TYPE_IMAGE = "image" as const;

const BizDeployNodeConfigFieldsProviderVolcEngineVOD = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldDomainMatchPattern = Form.useWatch([parentNamePath, "domainMatchPattern"], {
    form: formInst,
    preserve: true,
  });

  return (
    <>
      <Form.Item
        name={[parentNamePath, "spaceName"]}
        initialValue={initialValues.spaceName}
        label={t("workflow_node.deploy.form.volcengine_vod_space_name.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.volcengine_vod_space_name.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.volcengine_vod_space_name.placeholder")} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domainType"]}
        initialValue={initialValues.domainType}
        label={t("workflow_node.deploy.form.volcengine_vod_domain_type.label")}
        rules={[formRule]}
      >
        <Select
          options={[DOMAIN_TYPE_PLAY, DOMAIN_TYPE_IMAGE].map((s) => ({
            value: s,
            label: t(`workflow_node.deploy.form.volcengine_vod_domain_type.option.${s}.label`),
          }))}
          placeholder={t("workflow_node.deploy.form.volcengine_vod_domain_type.placeholder")}
        />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "domainMatchPattern"]}
        initialValue={initialValues.domainMatchPattern}
        label={t("workflow_node.deploy.form.shared_domain_match_pattern.label")}
        rules={[formRule]}
      >
        <Radio.Group
          options={[DOMAIN_MATCH_PATTERN_EXACT, DOMAIN_MATCH_PATTERN_WILDCARD, DOMAIN_MATCH_PATTERN_CERTSAN].map((s) => ({
            key: s,
            label: t(`workflow_node.deploy.form.shared_domain_match_pattern.option.${s}.label`),
            value: s,
          }))}
        />
      </Form.Item>

      <Show when={fieldDomainMatchPattern !== DOMAIN_MATCH_PATTERN_CERTSAN}>
        <Form.Item
          name={[parentNamePath, "domain"]}
          initialValue={initialValues.domain}
          label={t("workflow_node.deploy.form.volcengine_vod_domain.label")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.volcengine_vod_domain.placeholder")} />
        </Form.Item>
      </Show>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    spaceName: "",
    domainMatchPattern: DOMAIN_MATCH_PATTERN_EXACT,
    domainType: DOMAIN_TYPE_PLAY,
    domain: "",
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      spaceName: z.string().nonempty(t("workflow_node.deploy.form.volcengine_vod_space_name.placeholder")).nullish(),
      domainMatchPattern: z.string().nonempty(t("workflow_node.deploy.form.shared_domain_match_pattern.placeholder")).default(DOMAIN_MATCH_PATTERN_EXACT),
      domainType: z.literal([DOMAIN_TYPE_PLAY, DOMAIN_TYPE_IMAGE], t("workflow_node.deploy.form.volcengine_vod_domain_type.placeholder")),
      domain: z.string().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.domainMatchPattern) {
        switch (values.domainMatchPattern) {
          case DOMAIN_MATCH_PATTERN_EXACT:
          case DOMAIN_MATCH_PATTERN_WILDCARD:
            {
              if (!validDomainName(values.domain!, { allowWildcard: true })) {
                ctx.addIssue({
                  code: "custom",
                  message: t("common.errmsg.domain_invalid"),
                  path: ["domain"],
                });
              }
            }
            break;
        }
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderVolcEngineVOD, {
  getInitialValues,
  getSchema,
});

export default _default;
