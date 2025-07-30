import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { App, Card, Col, Row, Spin, Typography } from "antd";
import dayjs from "dayjs";

import { type WorkflowModel, initWorkflow } from "@/domain/workflow";
import { save as saveWorkflow } from "@/repository/workflow";
import { getErrMsg } from "@/utils/error";

const TEMPLATE_KEY_STANDARD = "standard" as const;
const TEMPLATE_KEY_CERTTEST = "certtest" as const;
const TEMPLATE_KEY_EMPTY = "empty" as const;
type TemplateKeys = typeof TEMPLATE_KEY_EMPTY | typeof TEMPLATE_KEY_CERTTEST | typeof TEMPLATE_KEY_STANDARD;

const WorkflowNew = () => {
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { notification } = App.useApp();

  const templateGridSpans = {
    xs: { flex: "100%" },
    md: { flex: "50%" },
    lg: { flex: "50%" },
    xl: { flex: "33.3333%" },
    xxl: { flex: "33.3333%" },
  };
  const [templateSelectKey, setTemplateSelectKey] = useState<TemplateKeys>();

  const [pending, setPending] = useState(false);

  const handleTemplateClick = async (key: TemplateKeys) => {
    if (pending) return;

    setTemplateSelectKey(key);

    try {
      let workflow: WorkflowModel;

      switch (key) {
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

      workflow.name = t("workflow.new.templates.default_name");
      workflow.description = t("workflow.new.templates.default_description", { date: dayjs().format("YYYY-MM-DD HH:mm") });
      workflow = await saveWorkflow(workflow);
      navigate(`/workflows/${workflow.id}`, { replace: true });
    } catch (err) {
      notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

      throw err;
    } finally {
      setPending(false);
      setTemplateSelectKey(undefined);
    }
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("workflow.new.title")}</h1>
        <p className="text-base text-gray-500">{t("workflow.new.subtitle")}</p>
      </div>

      <div className="container">
        <Typography.Text type="secondary">
          <div className="mb-4 text-xl">{t("workflow.new.templates.title")}</div>
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
    </div>
  );
};

export default WorkflowNew;
