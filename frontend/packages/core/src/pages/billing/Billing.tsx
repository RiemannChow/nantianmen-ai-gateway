import InsidePage from '@common/components/aoplatform/InsidePage'
import { App, Alert, Button, Card, Col, DatePicker, Descriptions, Form, Input, InputNumber, Row, Space, Table, Tag } from 'antd'
import { useRef, useState } from 'react'
import dayjs from 'dayjs'
import { useGlobalContext } from '@common/contexts/GlobalStateContext'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { $t } from '@common/locales'

interface PricingConfig {
  per_call: number
  per_1k_input_token: number
  per_1k_output_token: number
  per_instance_month: number
}

interface RebateInput {
  rebate: { [provider: string]: number }
}

interface ProviderBill {
  provider: string
  calls: number
  input_tokens: number
  output_tokens: number
  call_cost: number
  token_cost: number
  bill: number
  rebate: number
}

interface BillingOverview {
  start: number
  end: number
  calls: number
  input_tokens: number
  output_tokens: number
  instances: number
  pricing: PricingConfig
  call_cost: number
  token_cost: number
  instance_cost: number
  bill_total: number
  rebate_total: number
  net_total: number
  providers: ProviderBill[]
}

const fmt = (v: number) => (v == null ? '-' : v.toFixed(2))
const fmtInt = (v: number) => (v == null ? '-' : `${v}`)

export default function Billing() {
  const { fetchData } = useGlobalContext()
  const { message: msgApi } = App.useApp()
  const [pricingForm] = Form.useForm<PricingConfig>()
  const [rebateForm] = Form.useForm<{ items: { provider: string; rate: number }[] }>()
  const [pricingLoading, setPricingLoading] = useState(false)
  const [rebateLoading, setRebateLoading] = useState(false)
  const [overview, setOverview] = useState<BillingOverview | null>(null)
  const [overviewLoading, setOverviewLoading] = useState(false)
  const [range, setRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([
    dayjs().subtract(30, 'day'),
    dayjs()
  ])
  const rangeRef = useRef(range)

  rangeRef.current = range

  const loadPricing = () => {
    setPricingLoading(true)
    fetchData<BasicResponse<{ pricing: PricingConfig }>>('billing/pricing', { method: 'GET' })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          const p = resp.data.pricing || {}
          pricingForm.setFieldsValue({
            per_call: p.per_call || 0,
            per_1k_input_token: p.per_1k_input_token || 0,
            per_1k_output_token: p.per_1k_output_token || 0,
            per_instance_month: p.per_instance_month || 0
          })
        } else {
          msgApi.error(resp?.msg || $t(RESPONSE_TIPS.error))
        }
      })
      .finally(() => setPricingLoading(false))
  }

  const loadRebate = () => {
    setRebateLoading(true)
    fetchData<BasicResponse<{ rebate: { [provider: string]: number } }>>('billing/rebate', { method: 'GET' })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          const r = resp.data.rebate || {}
          const items = Object.keys(r).map((provider) => ({ provider, rate: r[provider] }))
          rebateForm.setFieldsValue({ items: items.length ? items : [{ provider: '', rate: 0 }] })
        } else {
          msgApi.error(resp?.msg || $t(RESPONSE_TIPS.error))
        }
      })
      .finally(() => setRebateLoading(false))
  }

  const savePricing = (values: PricingConfig) => {
    setPricingLoading(true)
    fetchData<BasicResponse<{ pricing: PricingConfig }>>('billing/pricing', {
      method: 'PUT',
      eoBody: {
        per_call: values.per_call || 0,
        per_1k_input_token: values.per_1k_input_token || 0,
        per_1k_output_token: values.per_1k_output_token || 0,
        per_instance_month: values.per_instance_month || 0
      }
    })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          msgApi.success($t('定价配置已保存'))
        } else {
          msgApi.error(resp?.msg || $t(RESPONSE_TIPS.error))
        }
      })
      .finally(() => setPricingLoading(false))
  }

  const saveRebate = (values: { items: { provider: string; rate: number }[] }) => {
    const rebate: { [provider: string]: number } = {}
    ;(values.items || []).forEach((it) => {
      const name = (it.provider || '').trim()
      if (name) rebate[name] = it.rate || 0
    })
    setRebateLoading(true)
    fetchData<BasicResponse<RebateInput>>('billing/rebate', {
      method: 'PUT',
      eoBody: { rebate }
    })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          msgApi.success($t('返佣配置已保存'))
          loadRebate()
        } else {
          msgApi.error(resp?.msg || $t(RESPONSE_TIPS.error))
        }
      })
      .finally(() => setRebateLoading(false))
  }

  const computeOverview = () => {
    const [s, e] = rangeRef.current
    const start = Math.floor(s.toDate().getTime() / 1000)
    const end = Math.floor(e.toDate().getTime() / 1000)
    setOverviewLoading(true)
    fetchData<BasicResponse<{ billing: BillingOverview }>>('billing/overview', {
      method: 'GET',
      eoParams: { start, end }
    })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          setOverview(resp.data.billing)
        } else {
          msgApi.error(resp?.msg || $t(RESPONSE_TIPS.error))
        }
      })
      .finally(() => setOverviewLoading(false))
  }

  // 生成示例用量数据：在全新安装（无真实数据面流量）时，一键灌入有代表性的
  // 用量，使 FinOps 套件（账单/分析/成本优化）有数据可看；随后立即核算。
  const seedSample = () => {
    setOverviewLoading(true)
    fetchData<BasicResponse<{ result: { status: string; days: number } }>>('metering/seed', {
      method: 'POST',
      eoParams: { days: 30 }
    })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          msgApi.success($t('已生成示例用量数据，正在核算'))
          computeOverview()
        } else {
          msgApi.error(resp?.msg || $t(RESPONSE_TIPS.error))
          setOverviewLoading(false)
        }
      })
      .catch(() => setOverviewLoading(false))
  }

  return (
    <InsidePage
      pageTitle={$t('计费与返佣')}
      showBorder={false}
      contentClassName="pr-PAGE_INSIDE_X"
      scrollPage={false}
      description={$t('配置单价与模型供应商渠道返佣，并基于用量数据核算月度账单与返佣（FinOps）')}
    >
      <Alert
        className="mb-[16px]"
        type="info"
        showIcon
        message={$t('说明')}
        description={$t(
          '本模块在控制面内核算账单：调用数取自请求概况、Token 取自 AI 用量、实例数取已启用的数据面集群数。实际多租户云托管部署与第三方支付不在本模块范围。'
        )}
      />

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title={$t('定价配置')} loading={pricingLoading} className="mb-[16px]">
            <Form
              form={pricingForm}
              layout="vertical"
              initialValues={{ per_call: 0, per_1k_input_token: 0, per_1k_output_token: 0, per_instance_month: 0 }}
              onFinish={savePricing}
            >
              <Form.Item label={$t('每调用次单价')} name="per_call">
                <InputNumber className="w-full" min={0} step={0.001} precision={4} />
              </Form.Item>
              <Form.Item label={$t('每 1K 输入 Token 单价')} name="per_1k_input_token">
                <InputNumber className="w-full" min={0} step={0.001} precision={4} />
              </Form.Item>
              <Form.Item label={$t('每 1K 输出 Token 单价')} name="per_1k_output_token">
                <InputNumber className="w-full" min={0} step={0.001} precision={4} />
              </Form.Item>
              <Form.Item label={$t('每网关实例·月 单价')} name="per_instance_month">
                <InputNumber className="w-full" min={0} step={1} precision={2} />
              </Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={pricingLoading}>
                  {$t('保存定价')}
                </Button>
                <Button onClick={loadPricing} loading={pricingLoading}>
                  {$t('刷新')}
                </Button>
              </Space>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title={$t('渠道返佣配置')} loading={rebateLoading} className="mb-[16px]">
            <Form form={rebateForm} layout="vertical" onFinish={saveRebate}>
              <Form.List name="items">
                {(fields, { add, remove }) => (
                  <>
                    {fields.map((field) => (
                      <Space key={field.key} align="baseline" className="w-full mb-[8px]">
                        <Form.Item
                          label={$t('模型供应商')}
                          name={[field.name, 'provider']}
                          rules={[{ required: true, message: $t('请输入供应商标识') }]}
                        >
                          <InputWidth />
                        </Form.Item>
                        <Form.Item
                          label={$t('返佣单价/次')}
                          name={[field.name, 'rate']}
                          rules={[{ required: true, message: $t('请输入返佣单价') }]}
                        >
                          <InputNumber min={0} step={0.001} precision={4} />
                        </Form.Item>
                        <Button type="link" danger onClick={() => remove(field.name)}>
                          {$t('删除')}
                        </Button>
                      </Space>
                    ))}
                    <Button type="dashed" onClick={() => add({ provider: '', rate: 0 })} block>
                      {$t('添加供应商')}
                    </Button>
                  </>
                )}
              </Form.List>
              <Space className="mt-[12px]">
                <Button type="primary" htmlType="submit" loading={rebateLoading}>
                  {$t('保存返佣')}
                </Button>
                <Button onClick={loadRebate} loading={rebateLoading}>
                  {$t('刷新')}
                </Button>
              </Space>
            </Form>
          </Card>
        </Col>
      </Row>

      <Card
        title={$t('账单核算')}
        className="mb-[16px]"
        extra={
          <Space>
            <DatePicker.RangePicker
              value={range}
              onChange={(v) => v && setRange(v as [dayjs.Dayjs, dayjs.Dayjs])}
            />
            <Button loading={overviewLoading} onClick={seedSample}>
              {$t('生成示例用量数据')}
            </Button>
            <Button type="primary" loading={overviewLoading} onClick={computeOverview}>
              {$t('核算')}
            </Button>
          </Space>
        }
      >
        {overview && (
          <>
            <Descriptions column={3} bordered size="small" className="mb-[16px]">
              <Descriptions.Item label={$t('调用总数')}>{fmtInt(overview.calls)}</Descriptions.Item>
              <Descriptions.Item label={$t('输入 Token')}>{fmtInt(overview.input_tokens)}</Descriptions.Item>
              <Descriptions.Item label={$t('输出 Token')}>{fmtInt(overview.output_tokens)}</Descriptions.Item>
              <Descriptions.Item label={$t('网关实例数')}>{fmtInt(overview.instances)}</Descriptions.Item>
              <Descriptions.Item label={$t('调用费用')}>{fmt(overview.call_cost)}</Descriptions.Item>
              <Descriptions.Item label={$t('Token 费用')}>{fmt(overview.token_cost)}</Descriptions.Item>
              <Descriptions.Item label={$t('实例费用')}>{fmt(overview.instance_cost)}</Descriptions.Item>
              <Descriptions.Item label={$t('账单合计')}>
                <Tag color="blue">{fmt(overview.bill_total)}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label={$t('返佣合计')}>
                <Tag color="green">{fmt(overview.rebate_total)}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label={$t('净额（账单-返佣）')}>
                <Tag color="purple">{fmt(overview.net_total)}</Tag>
              </Descriptions.Item>
            </Descriptions>

            <Table<ProviderBill>
              rowKey="provider"
              size="small"
              pagination={false}
              dataSource={overview.providers || []}
              columns={[
                { title: $t('模型供应商'), dataIndex: 'provider', key: 'provider' },
                { title: $t('调用数'), dataIndex: 'calls', key: 'calls', render: fmtInt },
                { title: $t('输入 Token'), dataIndex: 'input_tokens', key: 'input_tokens', render: fmtInt },
                { title: $t('输出 Token'), dataIndex: 'output_tokens', key: 'output_tokens', render: fmtInt },
                { title: $t('调用费用'), dataIndex: 'call_cost', key: 'call_cost', render: fmt },
                { title: $t('Token 费用'), dataIndex: 'token_cost', key: 'token_cost', render: fmt },
                { title: $t('账单'), dataIndex: 'bill', key: 'bill', render: fmt },
                { title: $t('返佣'), dataIndex: 'rebate', key: 'rebate', render: fmt }
              ]}
            />
          </>
        )}
        {!overview && (
          <div className="text-[13px] text-[#8a8aa0]">{$t('选择时间范围并点击「核算」生成账单与返佣明细。')}</div>
        )}
      </Card>
    </InsidePage>
  )
}

// 小宽度的供应商输入框，避免在弹性布局里撑满整行
function InputWidth() {
  return <Input style={{ width: 200 }} placeholder={$t('如 openai / azure / 通义千问')} />
}
