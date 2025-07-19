import { CopyToClipboard } from "react-copy-to-clipboard";
import { useTranslation } from "react-i18next";
import { IconChevronDown, IconClipboard, IconThumbUp } from "@tabler/icons-react";
import { App, Button, Dropdown, Form, Input, Tooltip } from "antd";
import dayjs from "dayjs";
import { saveAs } from "file-saver";

import { archive as archiveCertificate } from "@/api/certificates";
import { CERTIFICATE_FORMATS, type CertificateFormatType, type CertificateModel } from "@/domain/certificate";

export type CertificateDetailProps = {
  className?: string;
  style?: React.CSSProperties;
  data: CertificateModel;
};

const CertificateDetail = ({ data, ...props }: CertificateDetailProps) => {
  const { t } = useTranslation();

  const { message } = App.useApp();

  const handleDownloadClick = async (format: CertificateFormatType) => {
    try {
      const res = await archiveCertificate(data.id, format);
      const bstr = atob(res.data.fileBytes);
      const u8arr = Uint8Array.from(bstr, (ch) => ch.charCodeAt(0));
      const blob = new Blob([u8arr], { type: "application/zip" });
      saveAs(blob, `${data.id}-${data.subjectAltNames}.zip`);
    } catch (err) {
      console.error(err);
      message.warning(t("common.text.operation_failed"));
    }
  };

  return (
    <div {...props}>
      <Form layout="vertical">
        <Form.Item label={t("certificate.props.subject_alt_names")}>
          <Input value={data.subjectAltNames} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.issuer")}>
          <Input value={data.issuerOrg} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.validity")}>
          <Input
            value={`${dayjs(data.effectAt).format("YYYY-MM-DD HH:mm:ss")} ~ ${dayjs(data.expireAt).format("YYYY-MM-DD HH:mm:ss")}`}
            variant="filled"
            placeholder=""
          />
        </Form.Item>

        <Form.Item label={t("certificate.props.serial_number")}>
          <Input value={data.serialNumber} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item label={t("certificate.props.key_algorithm")}>
          <Input value={data.keyAlgorithm} variant="filled" placeholder="" />
        </Form.Item>

        <Form.Item>
          <div className="mb-2 flex w-full items-center justify-between">
            <label>{t("certificate.props.certificate")}</label>
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

        <Form.Item>
          <div className="mb-2 flex w-full items-center justify-between">
            <label>{t("certificate.props.private_key")}</label>
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
        <Dropdown
          menu={{
            items: [
              {
                key: "PEM",
                label: "PEM",
                extra: <IconThumbUp size="1.25em" />,
                onClick: () => handleDownloadClick(CERTIFICATE_FORMATS.PEM),
              },
              {
                key: "PFX",
                label: "PFX",
                onClick: () => handleDownloadClick(CERTIFICATE_FORMATS.PFX),
              },
              {
                key: "JKS",
                label: "JKS",
                onClick: () => handleDownloadClick(CERTIFICATE_FORMATS.JKS),
              },
            ],
          }}
        >
          <Button icon={<IconChevronDown size="1.25em" />} iconPosition="end" type="primary">
            {t("certificate.action.download")}
          </Button>
        </Dropdown>
      </div>
    </div>
  );
};

export default CertificateDetail;
