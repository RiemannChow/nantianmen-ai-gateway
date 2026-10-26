import { App, Button, Card, Descriptions, Input, Space, Tag, Upload, message } from 'antd'
import { useState } from 'react'
import { useGlobalContext } from '@common/contexts/GlobalStateContext'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { $t } from '@common/locales'
import { InboxOutlined } from '@ant-design/icons'

interface LicenseView {
  edition: string
  customer: string
  modules: string[]
  max_nodes: number
  issued_at: number
  expiry: number
  sla: string
  valid: boolean
  loaded: boolean
}

interface ModuleStatus {
  name: string
  label: string
  enabled: boolean
  reason?: string
}

function formatDate(unix: number): string {
  if (!unix || unix <= 0) return $t('永久有效')
  const d = new Date(unix * 1000)
  const pad = (n: number) => `${n}`.padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

function reasonText(reason?: string): string {
  switch (reason) {
    case 'expired':
      return $t('已过期')
    case 'not_included':
      return $t('未包含该模块')
    case 'no_license':
      return $t('未激活企业版')
    default:
      return $t('不可用')
  }
}

export default function License() {
  const { fetchData, $t: _t } = useGlobalContext()
  const { message: msgApi } = App.useApp()
  const [license, setLicense] = useState<LicenseView | null>(null)
  const [modules, setModules] = useState<ModuleStatus[]>([])
  const [loading, setLoading] = useState(false)
  const [activating, setActivating] = useState(false)
  const [payload, setPayload] = useState('')

  const load = () => {
    setLoading(true)
    Promise.all([
      fetchData<BasicResponse<{ license: LicenseView }>>('common/license', { method: 'GET' }),
      fetchData<BasicResponse<{ modules: { modules: ModuleStatus[] } }>>('license/modules', { method: 'GET' })
    ])
      .then(([licResp, modResp]) => {
        if (licResp?.code === STATUS_CODE.SUCCESS) setLicense(licResp.data.license)
        if (modResp?.code === STATUS_CODE.SUCCESS) setModules(modResp.data.modules.modules)
      })
      .finally(() => setLoading(false))
  }

  const activate = () => {
    const text = payload.trim()
    if (!text) {
      msgApi.warning($t('请粘贴或上传 License 内容'))
      return
    }
    setActivating(true)
    fetchData<BasicResponse<{ license: LicenseView }>>('license/activate', {
      method: 'POST',
      eoBody: { payload: text }
    })
      .then((response) => {
        const { code, data, msg } = response
        if (code === STATUS_CODE.SUCCESS) {
          msgApi.success($t('License 激活成功'))
          setLicense(data.license)
          setPayload('')
          load()
        } else {
          msgApi.error(msg || $t(RESPONSE_TIPS.error))
        }
      })
      .finally(() => setActivating(false))
  }

  const onFile = (file: File) => {
    const reader = new FileReader()
    reader.onload = () => setPayload(`${reader.result}`)
    reader.readAsText(file)
    return false
  }

  const isEnterprise = license?.edition === 'enterprise' && license?.valid

  return (
    <div className="w-full h-full overflow-auto p-PAGE_INSIDE_X">
      <div className="max-w-[960px] mx-auto">
        <Card
          title={$t('授权管理')}
          className="mb-[16px]"
          extra={
            <Tag color={isEnterprise ? 'gold' : 'default'}>
              {isEnterprise ? $t('企业版') : $t('社区版')}
            </Tag>
          }
          loading={loading && !license}
        >
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label={$t('版本')}>
              {isEnterprise ? $t('南天门 AI 网关 · 企业版') : $t('南天门 AI 网关 · 社区版（开源）')}
            </Descriptions.Item>
            <Descriptions.Item label={$t('授权客户')}>{license?.customer || '-'}</Descriptions.Item>
            <Descriptions.Item label={$t('签发时间')}>{formatDate(license?.issued_at || 0)}</Descriptions.Item>
            <Descriptions.Item label={$t('到期时间')}>{formatDate(license?.expiry || 0)}</Descriptions.Item>
            <Descriptions.Item label={$t('SLA')}>{license?.sla || '-'}</Descriptions.Item>
            <Descriptions.Item label={$t('网关实例上限')}>
              {license?.max_nodes ? license.max_nodes : $t('不限')}
            </Descriptions.Item>
            <Descriptions.Item label={$t('License 状态')} span={2}>
              {!license?.loaded ? (
                <Tag color="default">{$t('未激活')}</Tag>
              ) : license.valid ? (
                <Tag color="success">{$t('有效')}</Tag>
              ) : (
                <Tag color="error">{$t('无效 / 签名校验失败')}</Tag>
              )}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        <Card title={$t('商业版能力')} className="mb-[16px]" loading={loading && modules.length === 0}>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-[12px]">
            {modules.map((m) => (
              <div
                key={m.name}
                className="flex items-center justify-between rounded-[8px] border border-[#2a2850] px-[14px] py-[12px]"
              >
                <div>
                  <div className="text-[14px] text-[#e6e6f0]">{m.label}</div>
                  {!m.enabled && <div className="text-[12px] text-[#ff9c6e] mt-[2px]">{reasonText(m.reason)}</div>}
                </div>
                {m.enabled ? (
                  <Tag color="success">{$t('已授权')}</Tag>
                ) : (
                  <Tag color="warning">{$t('升级企业版')}</Tag>
                )}
              </div>
            ))}
          </div>
        </Card>

        <Card title={$t('激活 / 更换 License')}>
          <Space direction="vertical" className="w-full">
            <Upload beforeUpload={onFile} showUploadList={false} accept=".json,.txt">
              <Button icon={<InboxOutlined />}>{$t('上传 License 文件')}</Button>
            </Upload>
            <Input.TextArea
              rows={8}
              value={payload}
              onChange={(e) => setPayload(e.target.value)}
              placeholder={$t('粘贴由厂商签发的 License JSON 内容')}
            />
            <Button type="primary" loading={activating} onClick={activate}>
              {$t('激活')}
            </Button>
          </Space>
        </Card>
      </div>
    </div>
  )
}
