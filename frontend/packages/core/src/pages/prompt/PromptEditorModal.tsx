import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { useFetch } from '@common/hooks/http'
import { $t } from '@common/locales/index.ts'
import PromptEditor from '@common/components/aoplatform/prompt-editor/PromptEditor'
import { App, Form, Input, Select } from 'antd'
import { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import { PromptManagementHandle, PromptTemplate } from './type'

const { TextArea } = Input

const PromptEditorModal = forwardRef<PromptManagementHandle, { template?: PromptTemplate | null }>(
  ({ template }, ref) => {
    const [form] = Form.useForm()
    const { message } = App.useApp()
    const { fetchData } = useFetch()
    const [content, setContent] = useState<string>(template?.content || '')

    useEffect(() => {
      if (template) {
        form.setFieldsValue({
          name: template.name,
          tags: template.tags || [],
          variables: template.variables || []
        })
        setContent(template.content || '')
      } else {
        form.resetFields()
        setContent('')
      }
    }, [template])

    useImperativeHandle(ref, () => ({
      save: () => {
        return new Promise<boolean | string>((resolve) => {
          form
            .validateFields()
            .then((values) => {
              const body = {
                id: template?.id || '',
                name: values.name,
                tags: values.tags || [],
                variables: values.variables || [],
                content
              }
              fetchData<BasicResponse<{ template: PromptTemplate }>>('ai/prompt', {
                method: 'POST',
                eoBody: body
              })
                .then((response) => {
                  const { code, msg } = response
                  if (code === STATUS_CODE.SUCCESS) {
                    message.success($t('保存成功'))
                    resolve(true)
                  } else {
                    message.error(msg || $t(RESPONSE_TIPS.error))
                    resolve(msg || $t(RESPONSE_TIPS.error))
                  }
                })
                .catch(() => {
                  message.error($t(RESPONSE_TIPS.error))
                  resolve($t(RESPONSE_TIPS.error))
                })
            })
            .catch(() => {
              resolve(false)
            })
        })
      }
    }))

    return (
      <Form form={form} layout="vertical">
        <Form.Item
          name="name"
          label={$t('名称')}
          rules={[{ required: true, message: $t('请输入名称') }]}
        >
          <Input placeholder={$t('请输入提示词名称')} />
        </Form.Item>
        <Form.Item name="tags" label={$t('标签')}>
          <Select mode="tags" placeholder={$t('输入后回车添加标签')} tokenSeparators={[',']} />
        </Form.Item>
        <Form.Item name="variables" label={$t('变量')}>
          <Select mode="tags" placeholder={$t('可插入的变量名，如 user_name')} tokenSeparators={[',']} />
        </Form.Item>
        <Form.Item label={$t('内容')} required>
          <PromptEditor
            value={content}
            onChange={(text: string) => setContent(text)}
            placeholder={$t('在此编写提示词内容，可使用变量块')}
          />
        </Form.Item>
      </Form>
    )
  }
)

export default PromptEditorModal
