import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { App, Card, Col, Form, Input, type InputRef, Row, Spin, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import ModalForm from "@/components/ModalForm";
import { type WorkflowModel, initWorkflow } from "@/domain/workflow";
import { useAntdForm } from "@/hooks";
import { save as saveWorkflow } from "@/repository/workflow";
import { getErrMsg } from "@/utils/error";

const TEMPLATE_KEY_STANDARD = "standard" as const;
const TEMPLATE_KEY_CERTTEST = "monitor" as const;
const TEMPLATE_KEY_EMPTY = "empty" as const;
type TemplateKeys = typeof TEMPLATE_KEY_EMPTY | typeof TEMPLATE_KEY_CERTTEST | typeof TEMPLATE_KEY_STANDARD;

const WorkflowNew = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { notification } = App.useApp();

  const templateGridSpans = {
    xs: { flex: "100%" },
    md: { flex: "100%" },
    lg: { flex: "50%" },
    xl: { flex: "33.3333%" },
    xxl: { flex: "33.3333%" },
  };
  const [templateSelectKey, setTemplateSelectKey] = useState<TemplateKeys>();

  const formSchema = z.object({
    name: z
      .string()
      .min(1, t("workflow.new.modal.form.name.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    description: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
    submit: submitForm,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: {
      name: "",
      description: "",
    },
    onSubmit: async (values) => {
      try {
        let workflow: WorkflowModel;

        switch (templateSelectKey) {
          case TEMPLATE_KEY_EMPTY:
            workflow = initWorkflow();
            break;

          case TEMPLATE_KEY_STANDARD:
            workflow = initWorkflow({ template: "standard" });
            break;

          case TEMPLATE_KEY_CERTTEST:
            workflow = initWorkflow({ template: "certtest" });
            break;

          default:
            throw "Invalid state: `templateSelectKey`";
        }

        workflow.name = values.name?.trim() ?? workflow.name;
        workflow.description = values.description?.trim() ?? workflow.description;
        workflow = await saveWorkflow(workflow);
        navigate(`/workflows/${workflow.id}`, { replace: true });
      } catch (err) {
        notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });
  const [formModalOpen, setFormModalOpen] = useState(false);

  useEffect(() => {
    if (formModalOpen) {
      setTimeout(() => inputRef.current?.focus({ cursor: "end" }), 1);
    } else {
      setTemplateSelectKey(undefined);
      formInst.resetFields();
    }
  }, [formModalOpen]);

  const inputRef = useRef<InputRef>(null);

  const handleTemplateClick = (key: TemplateKeys) => {
    setTemplateSelectKey(key);
    setFormModalOpen(true);
  };

  const handleModalOpenChange = (open: boolean) => {
    setFormModalOpen(open);
  };

  const handleModalFormFinish = () => {
    return submitForm();
  };

  return (
    <div>
      <Card styles={{ body: { padding: 0 } }}>
        <div className="px-6 py-4">
          <div className="mx-auto max-w-320">
            <h1>{t("workflow.new.title")}</h1>
            <p className="text-base text-gray-500">{t("workflow.new.subtitle")}</p>
          </div>
        </div>
      </Card>

      <div className="p-4">
        <div className="mx-auto max-w-320 px-2">
          <Typography.Text type="secondary">
            <div className="mt-4 mb-8 text-xl">{t("workflow.new.templates.title")}</div>
          </Typography.Text>

          <Row className="justify-stretch" gutter={[16, 16]}>
            <Col {...templateGridSpans}>
              <Card
                className="size-full"
                cover={<img className="min-h-[120px] object-contain" src="/imgs/workflow/tpl-standard.png" />}
                hoverable
                onClick={() => handleTemplateClick(TEMPLATE_KEY_STANDARD)}
              >
                <div className="flex w-full items-center gap-4">
                  <Card.Meta
                    className="grow"
                    title={t("workflow.new.templates.template.standard.title")}
                    description={t("workflow.new.templates.template.standard.description")}
                  />
                  <Spin spinning={templateSelectKey === TEMPLATE_KEY_STANDARD} />
                </div>
              </Card>
            </Col>

            <Col {...templateGridSpans}>
              <Card
                className="size-full"
                cover={<img className="min-h-[120px] object-contain" src="/imgs/workflow/tpl-certtest.png" />}
                hoverable
                onClick={() => handleTemplateClick(TEMPLATE_KEY_CERTTEST)}
              >
                <div className="flex w-full items-center gap-4">
                  <Card.Meta
                    className="grow"
                    title={t("workflow.new.templates.template.certtest.title")}
                    description={t("workflow.new.templates.template.certtest.description")}
                  />
                  <Spin spinning={templateSelectKey === TEMPLATE_KEY_CERTTEST} />
                </div>
              </Card>
            </Col>

            <Col {...templateGridSpans}>
              <Card
                className="size-full"
                cover={<img className="min-h-[120px] object-contain" src="/imgs/workflow/tpl-blank.png" />}
                hoverable
                onClick={() => handleTemplateClick(TEMPLATE_KEY_EMPTY)}
              >
                <div className="flex w-full items-center gap-4">
                  <Card.Meta
                    className="grow"
                    title={t("workflow.new.templates.template.empty.title")}
                    description={t("workflow.new.templates.template.empty.description")}
                  />
                  <Spin spinning={templateSelectKey === TEMPLATE_KEY_EMPTY} />
                </div>
              </Card>
            </Col>
          </Row>
        </div>

        <ModalForm
          {...formProps}
          autoFocus
          disabled={formPending}
          layout="vertical"
          form={formInst}
          modalProps={{ destroyOnHidden: true }}
          okText={t("common.button.submit")}
          open={formModalOpen}
          title={t(`workflow.new.modal.title`)}
          width={480}
          onFinish={handleModalFormFinish}
          onOpenChange={handleModalOpenChange}
        >
          <Form.Item name="name" label={t("workflow.new.modal.form.name.label")} rules={[formRule]}>
            <Input ref={inputRef} autoFocus placeholder={t("workflow.new.modal.form.name.placeholder")} />
          </Form.Item>

          <Form.Item name="description" label={t("workflow.new.modal.form.description.label")} rules={[formRule]}>
            <Input placeholder={t("workflow.new.modal.form.description.placeholder")} />
          </Form.Item>
        </ModalForm>
      </div>
    </div>
  );
};

export default WorkflowNew;
