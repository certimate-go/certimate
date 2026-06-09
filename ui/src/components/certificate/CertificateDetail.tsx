import { CopyToClipboard } from "react-copy-to-clipboard";
import { useTranslation } from "react-i18next";
import { IconClipboard, IconDownload } from "@tabler/icons-react";
import { App, Button, Form, Input, Tag, Tooltip } from "antd";
import dayjs from "dayjs";

import { type CertificateModel } from "@/domain/certificate";

import CertificateDownloadModal from "./CertificateDownloadModal";

export interface CertificateDetailProps {
  className?: string;
  style?: React.CSSProperties;
  data: CertificateModel;
}

const CertificateDetail = ({ data, ...props }: CertificateDetailProps) => {
  const { t } = useTranslation();

  const { message } = App.useApp();

  return (
    <div {...props}>
      <Form layout="vertical">
        <Form.Item label={t("certificate.props.subject_name")}>
          <Input value={data.subjectName} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.subject_alt_names")}>
          <Input value={data.subjectAltNames.split(";").join("; ")} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.issuer_name")}>
          <Input value={`${data.issuerName}`} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.issuer_org")}>
          <Input value={`${data.issuerOrg}`} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.validity")}>
          <Input
            value={`${dayjs(data.validityNotBefore).format("YYYY-MM-DD HH:mm:ss")} ~ ${dayjs(data.validityNotAfter).format("YYYY-MM-DD HH:mm:ss")}`}
            variant="filled"
            placeholder=""
            suffix={data.isRevoked ? <Tag color="error">{t("certificate.props.revoked")}</Tag> : <></>}
          />
        </Form.Item>

        <Form.Item label={t("certificate.props.serial_number")}>
          <Input value={data.serialNumber} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.key_algorithm")}>
          <Input value={data.keyAlgorithm} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.certificate")}>
          <div className="absolute -top-1.5 right-0 -translate-y-full">
            <Tooltip title={t("common.button.copy")}>
              <CopyToClipboard
                text={data.certificate}
                onCopy={() => {
                  message.success(t("common.text.copied"));
                }}
              >
                <Button size="small" type="text" icon={<IconClipboard size="1.25em" />}></Button>
              </CopyToClipboard>
            </Tooltip>
          </div>
          <Input.TextArea value={data.certificate} variant="filled" autoSize={{ minRows: 5, maxRows: 5 }} readOnly />
        </Form.Item>

        <Form.Item label={t("certificate.props.private_key")}>
          <div className="absolute -top-1.5 right-0 -translate-y-full">
            <Tooltip title={t("common.button.copy")}>
              <CopyToClipboard
                text={data.privateKey}
                onCopy={() => {
                  message.success(t("common.text.copied"));
                }}
              >
                <Button size="small" type="text" icon={<IconClipboard size="1.25em" />}></Button>
              </CopyToClipboard>
            </Tooltip>
          </div>
          <Input.TextArea value={data.privateKey} variant="filled" autoSize={{ minRows: 5, maxRows: 5 }} readOnly />
        </Form.Item>
      </Form>

      <div className="flex items-center justify-end">
        <CertificateDownloadModal
          data={data}
          trigger={
            <Button icon={<IconDownload size="1.25em" />} type="primary">
              {t("common.button.download")}
            </Button>
          }
        />
      </div>
    </div>
  );
};

export default CertificateDetail;
