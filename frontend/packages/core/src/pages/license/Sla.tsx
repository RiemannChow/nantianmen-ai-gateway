import InsidePage from '@common/components/aoplatform/InsidePage'
import { App, Alert, Card, Descriptions, Tag } from 'antd'
import { useEffect, useState } from 'react'
import { useGlobalContext } from '@common/contexts/GlobalStateContext'
import { useFetch } from '@common/hooks/http'
import { $t } from '@common/locales'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'

// 与后端 module/license 的 SLA 分级表保持一致（支持等级 / 可用性目标 / 响应时效）。
export type SLADetail = {
  level: string
  level_label: string
  availability: string
  response_time: string
  support: string
  description: string
}

const LEVEL_COLOR: Record<string, string> = {
  standard: 'default',
  gold: 'gold',
  platinum: 'purple',
  custom: 'blue'
}

const Sla = () => {
  const { message } = App.useApp()
  const { fetchData } = useFetch()
  const [sla, setSla] = useState<SLADetail | null>(null)
  const [enabled, setEnabled] = useState<boolean>(true)
  const [loading, setLoading] = useState<boolean>(true)

  useEffect(() => {
    fetchData<BasicResponse<{ modules: { modules: { name: string; enabled: boolean }[] } }>>('license/modules', { method: 'GET' })
      .then((response) => {
        const { code, data } = response
        if (code === STATUS_CODE.SUCCESS) {
          const mod = (data.modules.modules || []).find((m: any) => m.name === 'sla')
          setEnabled(!!(mod && mod.enabled))
        }
      })
      .catch(() => {
        /* 忽略，默认放行 */
      })
  }, [])

  useEffect(() => {
    if (!enabled) {
      setLoading(false)
      return
    }
    setLoading(true)
    fetchData<BasicResponse<{ sla: SLADetail }>>('sla', { method: 'GET' })
      .then((response) => {
        const { code, data, msg } = response
        if (code === STATUS_CODE.SUCCESS) {
          setSla(data.sla)
        } else {
          message.error(msg || $t(RESPONSE_TIPS.error))
        }
      })
      .catch(() => {
        /* 忽略 */
      })
      .finally(() => setLoading(false))
  }, [enabled])

  return (
    <InsidePage
      pageTitle={$t('SLA 支持')}
      showBorder={false}
      contentClassName="pr-PAGE_INSIDE_X"
      scrollPage={false}
      description={$t('企业版 SLA（服务等级协议）权益，包含支持等级、可用性目标与响应时效')}
    >
      {!enabled && (
        <Alert
          className="mb-[12px]"
          type="warning"
          showIcon
          message={$t('该能力属于企业版（SLA 支持）')}
          description={$t('当前 License 未包含 sla 模块，无法查看 SLA 权益。请升级企业版后使用。')}
        />
      )}

      {enabled && sla && (
        <Card loading={loading} title={$t('SLA 权益')}>
          <div className="mb-[16px]">
            <Tag color={LEVEL_COLOR[sla.level] || 'blue'} className="text-[14px] px-[12px] py-[4px]">
              {sla.level_label}
            </Tag>
          </div>
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label={$t('支持等级')}>{sla.level_label}</Descriptions.Item>
            <Descriptions.Item label={$t('可用性目标')}>{sla.availability}</Descriptions.Item>
            <Descriptions.Item label={$t('响应时效')}>{sla.response_time}</Descriptions.Item>
            <Descriptions.Item label={$t('支持方式')}>{sla.support}</Descriptions.Item>
            <Descriptions.Item label={$t('说明')}>{sla.description}</Descriptions.Item>
          </Descriptions>
          <Alert
            className="mt-[16px]"
            type="info"
            showIcon
            message={$t('可用性承诺基于控制面与数据面健康探测聚合')}
            description={$t('当前为权益展示。生产环境的实时可用性监控与 SLA 达标报表，将在后续版本基于流量日志与健康探测自动聚合。')}
          />
        </Card>
      )}
    </InsidePage>
  )
}

export default Sla
