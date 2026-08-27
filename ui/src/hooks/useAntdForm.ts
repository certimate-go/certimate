import { useMemo, useState } from "react";
import { useDeepCompareEffect } from "ahooks";
import { Form, type FormInstance, type FormProps } from "antd";

import { applyTrimmedFormValues, trimFormValues } from "@/utils/form";

import useAntdFormName from "./useAntdFormName";

export interface UseAntdFormOptions<T extends NonNullable<unknown> = any> {
  form?: FormInstance<T>;
  initialValues?: Partial<T | any> | (() => Partial<T | any> | Promise<Partial<T | any>>);
  name?: string;
  onSubmit?: (values: T) => unknown | Promise<unknown>;
}

export interface UseAntdFormReturns<T extends NonNullable<unknown> = any> {
  form: FormInstance<T>;
  formProps: Omit<FormProps<T>, "children">;
  formPending: boolean;
  submit: (values?: T) => Promise<unknown>;
}

/**
 * 生成并获取一个 antd 表单的实例、属性等。
 * 通常为配合 Form 组件使用，以减少样板代码。
 * @param {UseAntdFormOptions} options
 * @returns {UseAntdFormReturns}
 */
const useAntdForm = <T extends NonNullable<unknown> = any>({ form, initialValues, onSubmit, ...options }: UseAntdFormOptions<T>): UseAntdFormReturns<T> => {
  const formInst = form ?? Form["useForm"]()[0];
  const formName = useAntdFormName({ form: formInst, name: options.name });
  const [formInitialValues, setFormInitialValues] = useState<Partial<T>>();
  const [formPending, setFormPending] = useState(false);

  useDeepCompareEffect(() => {
    let unmounted = false;

    if (!initialValues) {
      return;
    }

    let p: Promise<Partial<T>>;
    if (typeof initialValues === "function") {
      p = Promise.resolve(initialValues());
    } else {
      p = Promise.resolve(initialValues);
    }

    p.then((res) => {
      if (!unmounted) {
        type FieldName = Parameters<FormInstance<T>["getFieldValue"]>[0];
        type FieldsValue = Parameters<FormInstance<T>["setFieldsValue"]>[0];

        const obj = { ...res };
        Object.keys(res).forEach((key) => {
          obj[key as keyof T] = formInst!.isFieldTouched(key as FieldName) ? formInst!.getFieldValue(key as FieldName) : res[key as keyof T];
        });

        setFormInitialValues(res);
        formInst!.setFieldsValue(obj as FieldsValue);
      }
    });

    return () => {
      unmounted = true;
    };
  }, [formInst, initialValues]);

  const wrappedFormInst = useMemo(() => {
    const originalSubmit = formInst.submit;
    return {
      ...formInst,
      submit: () => {
        applyTrimmedFormValues(formInst);
        return originalSubmit();
      },
    };
  }, [formInst]);

  const doFinish = () => {
    if (formPending) return Promise.reject(new Error("Form is pending"));

    setFormPending(true);

    const values = trimFormValues(formInst.getFieldsValue(true));

    return new Promise((resolve, reject) => {
      formInst
        .validateFields()
        .then(() => {
          resolve(
            Promise.resolve(onSubmit?.(values))
              .then((ret) => {
                setFormPending(false);
                return ret;
              })
              .catch((err) => {
                setFormPending(false);
                throw err;
              })
          );
        })
        .catch((err) => {
          setFormPending(false);
          reject(err);
        });
    });
  };

  const formProps: FormProps = {
    form: wrappedFormInst,
    initialValues: formInitialValues,
    name: formName,
    onFinish: doFinish,
  };

  return {
    form: wrappedFormInst,
    formProps: formProps,
    formPending: formPending,
    submit: () => {
      applyTrimmedFormValues(formInst);
      return doFinish();
    },
  };
};

export default useAntdForm;
