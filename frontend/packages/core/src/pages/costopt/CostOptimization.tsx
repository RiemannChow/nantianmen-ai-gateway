import InsidePage from '@common/components/aoplatform/InsidePage'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  DatePicker,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Row,
  Space,
  Table,
  Tabs,
  Tag
} from 'antd'
import { useRef, useState } from 'react'
import dayjs from 'dayjs'
import { useGlobalContext } from '@common/contexts/GlobalStateContext'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { $t } from '@common/locales'

interface ModuleStatus {
  name: string
  label: string
  enabled: boolean
  reason?: string
}

const fmt = (v: number) => (v == null ? '-' : v.toFixed(2))
const fmtInt = (v: number) => (v == null ? '-' : `${v}`)

export default function CostOptimization() {
  const { fetchData } = useGlobalContext()
  const { message: msgApi } = App.useApp()
  const [enabled, setEnabled] = useState(true)
  const [loading, setLoading] = useState(false)

  const [budgetForm] = Form.useForm<{ monthly_budget: number; alert_threshold: number }>()
  const [providersForm] = Form.useForm<{ items: ProviderCostItem[] }>()
  const [subsForm] = Form.useForm<{ items: { from: string; to: string }[] }>()
  const [judgeForm] = Form.useForm<{ base_url: string; api_key: string; model: string }>()
  const [evaluateForm] = Form.useForm<{ model: string; prompts: string; criteria: string; judge_model: string }>()
  const [notifyForm] = Form.useForm<{
    webhook_enabled: boolean
    webhook_url: string
    webhook_method: string
    webhook_headers: string
    webhook_template: string
    email_enabled: boolean
    email_to: string
    email_subject: string
    email_body: string
  }>()
  const [consumerForm] = Form.useForm<{ items: ConsumerBudgetRow[] }>()

  const [budgetStatus, setBudgetStatus] = useState<any>(null)
  const [analytics, setAnalytics] = useState<any>(null)
  const [recommend, setRecommend] = useState<any>(null)
  const [anomalies, setAnomalies] = useState<any>(null)
  const [evaluation, setEvaluation] = useState<any>(null)
  const [notify, setNotify] = useState<any>(null)
  const [consumerBudgets, setConsumerBudgets] = useState<any>(null)
  const [consumerStatus, setConsumerStatus] = useState<Record<string, any>>({})

  const [range, setRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([dayjs().subtract(30, 'day'), dayjs()])
  const rangeRef = useRef(range)
  rangeRef.current = range

  const loadEnabled = () => {
    fetchData<BasicResponse<{ modules: { modules: ModuleStatus[] } }>>('license/modules', { method: 'GET' })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          const m = (resp.data.modules?.modules || []).find((x) => x.name === 'cost_optimization')
          setEnabled(!!(m && m.enabled))
        }
      })
      .catch(() => setEnabled(true))
  }

  const api = <T,>(url: string, opts: any) =>
    fetchData<BasicResponse<T>>(url, opts).then((resp) => {
      if (resp?.code === STATUS_CODE.SUCCESS) return resp.data as T
      throw new Error(resp?.msg || $t(RESPONSE_TIPS.error))
    })

  const toast = (ok: boolean, msg: string) =>
    ok ? msgApi.success(msg) : msgApi.error(msg)

  // ---- budget ----
  const loadBudget = () => {
    setLoading(true)
    api<{ budget: { monthly_budget: number; alert_threshold: number } }>('costopt/budget', { method: 'GET' })
      .then((d) =>
        budgetForm.setFieldsValue({
          monthly_budget: d.budget?.monthly_budget || 0,
          alert_threshold: d.budget?.alert_threshold || 0.8
        })
      )
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
    api<{ status: any }>('costopt/budget/status', { method: 'GET' })
      .then((d) => setBudgetStatus(d.status))
      .catch(() => {})
  }
  const saveBudget = (v: any) => {
    setLoading(true)
    api<{ budget: any }>('costopt/budget', {
      method: 'PUT',
      eoBody: { monthly_budget: v.monthly_budget || 0, alert_threshold: v.alert_threshold || 0 }
    })
      .then(() => {
        toast(true, $t('预算已保存'))
        loadBudget()
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // ---- providers ----
  const loadProviders = () => {
    setLoading(true)
    api<{ providers: { [k: string]: ProviderCost } }>('costopt/providers', { method: 'GET' })
      .then((d) => {
        const items = Object.keys(d.providers || {}).map((provider) => ({
          provider,
          per_call: d.providers[provider].per_call || 0,
          per_1k_input_token: d.providers[provider].per_1k_input_token || 0,
          per_1k_output_token: d.providers[provider].per_1k_output_token || 0
        }))
        providersForm.setFieldsValue({ items: items.length ? items : [{ provider: '', per_call: 0, per_1k_input_token: 0, per_1k_output_token: 0 }] })
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const saveProviders = (v: any) => {
    const providers: { [k: string]: ProviderCost } = {}
    ;(v.items || []).forEach((it: any) => {
      const name = (it.provider || '').trim()
      if (name) providers[name] = { per_call: it.per_call || 0, per_1k_input_token: it.per_1k_input_token || 0, per_1k_output_token: it.per_1k_output_token || 0 }
    })
    setLoading(true)
    api<{ providers: any }>('costopt/providers', { method: 'PUT', eoBody: { providers } })
      .then(() => {
        toast(true, $t('模型成本画像已保存'))
        loadProviders()
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // ---- substitutes ----
  const loadSubs = () => {
    setLoading(true)
    api<{ substitutes: { [k: string]: string } }>('costopt/substitutes', { method: 'GET' })
      .then((d) => {
        const items = Object.keys(d.substitutes || {}).map((from) => ({ from, to: d.substitutes[from] }))
        subsForm.setFieldsValue({ items: items.length ? items : [{ from: '', to: '' }] })
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const saveSubs = (v: any) => {
    const substitutes: { [k: string]: string } = {}
    ;(v.items || []).forEach((it: any) => {
      const from = (it.from || '').trim()
      const to = (it.to || '').trim()
      if (from && to) substitutes[from] = to
    })
    setLoading(true)
    api<{ substitutes: any }>('costopt/substitutes', { method: 'PUT', eoBody: { substitutes } })
      .then(() => {
        toast(true, $t('替代策略已保存'))
        loadSubs()
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // ---- judge ----
  const loadJudge = () => {
    setLoading(true)
    api<{ judge: { base_url: string; api_key: string; model: string } }>('costopt/judge', { method: 'GET' })
      .then((d) => judgeForm.setFieldsValue({ base_url: d.judge?.base_url || '', api_key: d.judge?.api_key || '', model: d.judge?.model || '' }))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const saveJudge = (v: any) => {
    setLoading(true)
    api<{ judge: any }>('costopt/judge', {
      method: 'PUT',
      eoBody: { base_url: v.base_url || '', api_key: v.api_key || '', model: v.model || '' }
    })
      .then(() => toast(true, $t('裁判模型已保存')))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // ---- analytics / recommend / anomalies ----
  const windowParams = () => {
    const [s, e] = rangeRef.current
    return { start: Math.floor(s.toDate().getTime() / 1000), end: Math.floor(e.toDate().getTime() / 1000) }
  }
  const loadAnalytics = () => {
    setLoading(true)
    api<{ analytics: any }>('costopt/analytics', { method: 'GET', eoParams: windowParams() })
      .then((d) => setAnalytics(d.analytics))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const loadRecommend = () => {
    setLoading(true)
    api<{ recommend: any }>('costopt/recommend', { method: 'GET', eoParams: windowParams() })
      .then((d) => setRecommend(d.recommend))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const loadAnomalies = () => {
    setLoading(true)
    api<{ anomalies: any }>('costopt/anomalies', { method: 'GET', eoParams: windowParams() })
      .then((d) => setAnomalies(d.anomalies))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // 生成示例用量数据：在全新安装（无真实数据面流量）时，一键灌入有代表性的
  // 用量，使 FinOps 套件（账单/分析/成本优化）有数据可看；随后刷新当前视图。
  const seedSample = () => {
    setLoading(true)
    api<{ result: { status: string; days: number } }>('metering/seed', { method: 'POST', eoParams: { days: 30 } })
      .then(() => {
        msgApi.success($t('已生成示例用量数据'))
        loadAnalytics()
        loadRecommend()
        loadAnomalies()
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // ---- evaluate ----
  const runEvaluate = (v: any) => {
    const prompts = (v.prompts || '')
      .split('\n')
      .map((s: string) => s.trim())
      .filter(Boolean)
    if (prompts.length === 0) {
      msgApi.warning($t('请输入至少一个评测提示词'))
      return
    }
    setLoading(true)
    api<{ evaluation: any }>('costopt/evaluate', {
      method: 'POST',
      eoBody: {
        model: v.model || '',
        prompts,
        criteria: v.criteria || $t('正确性、相关性、简洁性（1-10 分）'),
        judge_model: v.judge_model || ''
      }
    })
      .then((d) => setEvaluation(d.evaluation))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // initial load
  if (enabled && !budgetStatus && !analytics) {
    loadEnabled()
    loadBudget()
    loadProviders()
    loadSubs()
    loadJudge()
  }

  // ---- notification delivery ----
  const loadNotify = () => {
    setLoading(true)
    api<{ notify: any }>('costopt/notify', { method: 'GET' })
      .then((d) => {
        const n = d.notify || {}
        const wh = n.webhook || {}
        const em = n.email || {}
        notifyForm.setFieldsValue({
          webhook_enabled: !!wh.enabled,
          webhook_url: wh.url || '',
          webhook_method: wh.method || 'POST',
          webhook_headers: Object.keys(wh.headers || {})
            .map((k) => `${k}: ${wh.headers[k]}`)
            .join('\n'),
          webhook_template: wh.template || '',
          email_enabled: !!em.enabled,
          email_to: (em.to || []).join('\n'),
          email_subject: em.subject_template || '',
          email_body: em.body_template || ''
        })
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const saveNotify = (v: any) => {
    const headers: { [k: string]: string } = {}
    ;(v.webhook_headers || '')
      .split('\n')
      .map((s: string) => s.trim())
      .filter(Boolean)
      .forEach((line: string) => {
        const idx = line.indexOf(':')
        if (idx > 0) headers[line.slice(0, idx).trim()] = line.slice(idx + 1).trim()
      })
    const payload = {
      webhook: {
        enabled: !!v.webhook_enabled,
        url: v.webhook_url || '',
        method: (v.webhook_method || 'POST').toUpperCase(),
        headers,
        template: v.webhook_template || ''
      },
      email: {
        enabled: !!v.email_enabled,
        to: (v.email_to || '')
          .split('\n')
          .map((s: string) => s.trim())
          .filter(Boolean),
        subject_template: v.email_subject || '',
        body_template: v.email_body || ''
      }
    }
    setLoading(true)
    api<{ notify: any }>('costopt/notify', { method: 'PUT', eoBody: payload })
      .then(() => toast(true, $t('通知配置已保存')))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const sendTestNotification = (channel: string) => {
    setLoading(true)
    api<{}>('costopt/notify/test', { method: 'POST', eoParams: { channel } })
      .then(() => toast(true, $t('测试告警已发送，请检查对应通道')))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // ---- per-consumer / per-key budget ----
  const loadConsumerBudgets = () => {
    setLoading(true)
    api<{ consumer_budgets: any }>('costopt/consumer_budgets', { method: 'GET' })
      .then((d) => {
        const items = Object.keys(d.consumer_budgets?.items || {}).map((consumer) => ({
          consumer,
          ...d.consumer_budgets.items[consumer]
        }))
        consumerForm.setFieldsValue({
          items: items.length ? items : [{ consumer: '', monthly_budget: 0, alert_threshold: 0.8, qps: 10, enabled: false, enforce: false }]
        })
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const saveConsumerBudgets = (v: any) => {
    const items: { [k: string]: any } = {}
    ;(v.items || []).forEach((it: any) => {
      const name = (it.consumer || '').trim()
      if (name)
        items[name] = {
          monthly_budget: it.monthly_budget || 0,
          alert_threshold: it.alert_threshold || 0.8,
          qps: it.qps || 10,
          enabled: !!it.enabled,
          enforce: !!it.enforce
        }
    })
    setLoading(true)
    api<{ consumer_budgets: any }>('costopt/consumer_budgets', { method: 'PUT', eoBody: { items } })
      .then(() => {
        toast(true, $t('消费者预算已保存'))
        loadConsumerBudgets()
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const loadConsumerStatus = (consumer: string) => {
    api<{ status: any }>(`costopt/consumer/${encodeURIComponent(consumer)}/status`, { method: 'GET' })
      .then((d) => setConsumerStatus((prev) => ({ ...prev, [consumer]: d.status })))
      .catch((e) => msgApi.error(e.message))
  }
  const publishConsumerQuota = (consumer: string) => {
    setLoading(true)
    api<{}>(`costopt/consumer/${encodeURIComponent(consumer)}/quota`, { method: 'POST' })
      .then(() => toast(true, $t('配额策略已下发（实际限流由数据面执行）')))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  const locked = (
    <Alert
      className="mb-[16px]"
      type="warning"
      showIcon
      message={$t('该能力属于企业版（AI 成本优化与增值分析）')}
      description={$t('当前 License 未包含 cost_optimization 模块，无法使用成本优化与增值分析。请升级企业版后使用。')}
    />
  )

  return (
    <InsidePage
      pageTitle={$t('AI 成本优化与增值分析')}
      showBorder={false}
      contentClassName="pr-PAGE_INSIDE_X"
      scrollPage={false}
      description={$t('成本优化建议、预算告警、用量分析、异常检测与模型效果评测（LLM-as-judge）——南天门 AI 网关的 FinOps 与增值能力')}
    >
      {!enabled && locked}

      <Tabs
        defaultActiveKey="budget"
        onChange={(key) => {
          if (key === 'analytics') loadAnalytics()
          if (key === 'recommend') loadRecommend()
          if (key === 'anomalies') loadAnomalies()
          if (key === 'notify') loadNotify()
          if (key === 'consumer') loadConsumerBudgets()
        }}
        items={[
          {
            key: 'budget',
            label: $t('预算与告警'),
            children: (
              <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                  <Card title={$t('月度预算')} loading={loading}>
                    <Form form={budgetForm} layout="vertical" initialValues={{ monthly_budget: 0, alert_threshold: 0.8 }} onFinish={saveBudget}>
                      <Form.Item label={$t('月度预算')} name="monthly_budget">
                        <InputNumber className="w-full" min={0} step={100} precision={2} />
                      </Form.Item>
                      <Form.Item label={$t('预警阈值（占预算比例）')} name="alert_threshold">
                        <InputNumber className="w-full" min={0} max={1} step={0.05} precision={2} />
                      </Form.Item>
                      <Space>
                        <Button type="primary" htmlType="submit" loading={loading}>{$t('保存')}</Button>
                        <Button onClick={loadBudget}>{$t('刷新')}</Button>
                      </Space>
                    </Form>
                  </Card>
                </Col>
                <Col xs={24} lg={12}>
                  <Card title={$t('预算状态')} loading={loading}>
                    {budgetStatus && (
                      <Descriptions column={1} bordered size="small">
                        <Descriptions.Item label={$t('是否配置')}>
                          {budgetStatus.configured ? $t('是') : $t('否')}
                        </Descriptions.Item>
                        <Descriptions.Item label={$t('月度预算')}>{fmt(budgetStatus.monthly_budget)}</Descriptions.Item>
                        <Descriptions.Item label={$t('本月已花费')}>{fmt(budgetStatus.month_spend)}</Descriptions.Item>
                        <Descriptions.Item label={$t('占比')}>
                          {(budgetStatus.ratio * 100).toFixed(0)}%
                        </Descriptions.Item>
                        <Descriptions.Item label={$t('状态')}>
                          <Tag color={budgetStatus.level === 'exceeded' ? 'red' : budgetStatus.level === 'warning' ? 'orange' : 'green'}>
                            {budgetStatus.level}
                          </Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label={$t('说明')}>{budgetStatus.message}</Descriptions.Item>
                      </Descriptions>
                    )}
                    {!budgetStatus && <span className="text-[13px] text-[#8a8aa0]">{$t('请先配置并保存月度预算')}</span>}
                  </Card>
                </Col>
              </Row>
            )
          },
          {
            key: 'profiles',
            label: $t('模型成本画像'),
            children: (
              <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                  <Card title={$t('各模型（供应商）单价')} loading={loading}>
                    <Form form={providersForm} layout="vertical" onFinish={saveProviders}>
                      <Form.List name="items">
                        {(fields, { add, remove }) => (
                          <>
                            {fields.map((field) => (
                              <Space key={field.key} align="baseline" className="w-full mb-[8px]" wrap>
                                <Form.Item label={$t('供应商')} name={[field.name, 'provider']} rules={[{ required: true }]}>
                                  <Input style={{ width: 160 }} placeholder="openai" />
                                </Form.Item>
                                <Form.Item label={$t('每调用次')} name={[field.name, 'per_call']}>
                                  <InputNumber min={0} step={0.001} precision={4} />
                                </Form.Item>
                                <Form.Item label={$t('每1K输入')} name={[field.name, 'per_1k_input_token']}>
                                  <InputNumber min={0} step={0.001} precision={4} />
                                </Form.Item>
                                <Form.Item label={$t('每1K输出')} name={[field.name, 'per_1k_output_token']}>
                                  <InputNumber min={0} step={0.001} precision={4} />
                                </Form.Item>
                                <Button type="link" danger onClick={() => remove(field.name)}>{$t('删除')}</Button>
                              </Space>
                            ))}
                            <Button type="dashed" onClick={() => add({ provider: '', per_call: 0, per_1k_input_token: 0, per_1k_output_token: 0 })} block>
                              {$t('添加供应商')}
                            </Button>
                          </>
                        )}
                      </Form.List>
                      <Space className="mt-[12px]">
                        <Button type="primary" htmlType="submit" loading={loading}>{$t('保存')}</Button>
                        <Button onClick={loadProviders}>{$t('刷新')}</Button>
                      </Space>
                    </Form>
                  </Card>
                </Col>
                <Col xs={24} lg={12}>
                  <Card title={$t('可替代关系（from → to）')} loading={loading}>
                    <Form form={subsForm} layout="vertical" onFinish={saveSubs}>
                      <Form.List name="items">
                        {(fields, { add, remove }) => (
                          <>
                            {fields.map((field) => (
                              <Space key={field.key} align="baseline" className="w-full mb-[8px]">
                                <Form.Item label={$t('原供应商')} name={[field.name, 'from']} rules={[{ required: true }]}>
                                  <Input style={{ width: 160 }} placeholder="gpt-4o" />
                                </Form.Item>
                                <Form.Item label={$t('替代供应商')} name={[field.name, 'to']} rules={[{ required: true }]}>
                                  <Input style={{ width: 160 }} placeholder="gpt-4o-mini" />
                                </Form.Item>
                                <Button type="link" danger onClick={() => remove(field.name)}>{$t('删除')}</Button>
                              </Space>
                            ))}
                            <Button type="dashed" onClick={() => add({ from: '', to: '' })} block>
                              {$t('添加替代关系')}
                            </Button>
                          </>
                        )}
                      </Form.List>
                      <Space className="mt-[12px]">
                        <Button type="primary" htmlType="submit" loading={loading}>{$t('保存')}</Button>
                        <Button onClick={loadSubs}>{$t('刷新')}</Button>
                      </Space>
                    </Form>
                  </Card>
                </Col>
              </Row>
            )
          },
          {
            key: 'analytics',
            label: $t('用量分析'),
            children: (
              <Card
                title={$t('供应商支出分析')}
                loading={loading}
                extra={
                  <Space>
                    <DatePicker.RangePicker value={range} onChange={(v) => v && setRange(v as [dayjs.Dayjs, dayjs.Dayjs])} />
                    <Button loading={loading} onClick={seedSample}>{$t('生成示例用量数据')}</Button>
                    <Button type="primary" loading={loading} onClick={loadAnalytics}>{$t('分析')}</Button>
                  </Space>
                }
              >
                {analytics && (
                  <>
                    <Descriptions column={3} bordered size="small" className="mb-[16px]">
                      <Descriptions.Item label={$t('总支出')}>{fmt(analytics.total_spend)}</Descriptions.Item>
                      <Descriptions.Item label={$t('总调用')}>{fmtInt(analytics.total_calls)}</Descriptions.Item>
                      <Descriptions.Item label={$t('总 Token')}>{fmtInt(analytics.total_tokens)}</Descriptions.Item>
                    </Descriptions>
                    <Table
                      rowKey="provider"
                      size="small"
                      pagination={false}
                      dataSource={analytics.providers || []}
                      columns={[
                        { title: $t('供应商'), dataIndex: 'provider', key: 'provider' },
                        { title: $t('调用数'), dataIndex: 'calls', key: 'calls', render: fmtInt },
                        { title: $t('输入Token'), dataIndex: 'input_tokens', key: 'input_tokens', render: fmtInt },
                        { title: $t('输出Token'), dataIndex: 'output_tokens', key: 'output_tokens', render: fmtInt },
                        { title: $t('支出'), dataIndex: 'spend', key: 'spend', render: fmt },
                        { title: $t('占比'), dataIndex: 'share', key: 'share', render: (v: number) => `${((v || 0) * 100).toFixed(0)}%` }
                      ]}
                    />
                  </>
                )}
                {!analytics && <span className="text-[13px] text-[#8a8aa0]">{$t('选择时间范围并点击「分析」')}</span>}
              </Card>
            )
          },
          {
            key: 'recommend',
            label: $t('优化建议'),
            children: (
              <Card
                title={$t('成本优化建议')}
                loading={loading}
                extra={
                  <Space>
                    <DatePicker.RangePicker value={range} onChange={(v) => v && setRange(v as [dayjs.Dayjs, dayjs.Dayjs])} />
                    <Button type="primary" loading={loading} onClick={loadRecommend}>{$t('生成建议')}</Button>
                  </Space>
                }
              >
                {recommend && (
                  <>
                    <Alert className="mb-[12px]" type="success" showIcon message={$t('预计可节省合计')} description={fmt(recommend.total_projected_savings)} />
                    {(recommend.recommendations || []).map((r: any, idx: number) => (
                      <Card key={idx} size="small" className="mb-[8px]" title={r.title}>
                        <div className="text-[13px] text-[#c8c8d8] mb-[6px]">{r.detail}</div>
                        <Tag color="blue">{$t('类型')}: {r.type}</Tag>
                        {r.projected_savings > 0 && <Tag color="green">{$t('预计节省')}: {fmt(r.projected_savings)}</Tag>}
                      </Card>
                    ))}
                    {(recommend.recommendations || []).length === 0 && (
                      <span className="text-[13px] text-[#8a8aa0]">{$t('当前暂无具体优化建议，配置模型成本画像（各供应商单价）后效果更佳。')}</span>
                    )}
                  </>
                )}
                {!recommend && <span className="text-[13px] text-[#8a8aa0]">{$t('点击「生成建议」获取成本优化方案')}</span>}
              </Card>
            )
          },
          {
            key: 'anomalies',
            label: $t('异常检测'),
            children: (
              <Card
                title={$t('异常与风险')}
                loading={loading}
                extra={
                  <Space>
                    <DatePicker.RangePicker value={range} onChange={(v) => v && setRange(v as [dayjs.Dayjs, dayjs.Dayjs])} />
                    <Button type="primary" loading={loading} onClick={loadAnomalies}>{$t('检测')}</Button>
                  </Space>
                }
              >
                {anomalies && (anomalies.anomalies || []).map((a: any, idx: number) => (
                  <Alert
                    key={idx}
                    className="mb-[8px]"
                    type={a.severity === 'critical' ? 'error' : a.severity === 'warning' ? 'warning' : 'info'}
                    showIcon
                    message={a.title}
                    description={a.detail}
                  />
                ))}
                {anomalies && (anomalies.anomalies || []).length === 0 && (
                  <Alert type="success" showIcon message={$t('未检测到异常')} description={$t('当前时间范围内未发现预算或集中度风险。')} />
                )}
                {!anomalies && <span className="text-[13px] text-[#8a8aa0]">{$t('点击「检测」查看异常与风险')}</span>}
              </Card>
            )
          },
          {
            key: 'judge',
            label: $t('模型效果评测'),
            children: (
              <Row gutter={[16, 16]}>
                <Col xs={24} lg={12}>
                  <Card title={$t('裁判模型配置（OpenAI 兼容）')} loading={loading}>
                    <Form form={judgeForm} layout="vertical" onFinish={saveJudge}>
                      <Form.Item label={$t('接口地址 base_url')} name="base_url">
                        <Input placeholder="https://api.openai.com/v1" />
                      </Form.Item>
                      <Form.Item label={$t('API Key')} name="api_key">
                        <Input.Password placeholder={$t('选填，取决于供应商')} />
                      </Form.Item>
                      <Form.Item label={$t('裁判模型 judge_model')} name="model">
                        <Input placeholder="gpt-4o" />
                      </Form.Item>
                      <Space>
                        <Button type="primary" htmlType="submit" loading={loading}>{$t('保存')}</Button>
                        <Button onClick={loadJudge}>{$t('刷新')}</Button>
                      </Space>
                    </Form>
                  </Card>
                </Col>
                <Col xs={24} lg={12}>
                  <Card title={$t('发起评测（LLM-as-judge）')} loading={loading}>
                    <Form form={evaluateForm} layout="vertical" onFinish={runEvaluate}>
                      <Form.Item label={$t('被评测模型（标签）')} name="model">
                        <Input placeholder="gpt-4o-mini" />
                      </Form.Item>
                      <Form.Item label={$t('评测提示词（每行一个）')} name="prompts" rules={[{ required: true }]}>
                        <Input.TextArea rows={4} placeholder={'解释什么是向量数据库\n用一句话介绍 Kubernetes'} />
                      </Form.Item>
                      <Form.Item label={$t('评测标准')} name="criteria">
                        <Input.TextArea rows={2} />
                      </Form.Item>
                      <Form.Item label={$t('裁判模型（可选，默认用配置项）')} name="judge_model">
                        <Input placeholder={$t('留空则用上方配置')} />
                      </Form.Item>
                      <Button type="primary" htmlType="submit" loading={loading}>{$t('开始评测')}</Button>
                    </Form>
                    {evaluation && (
                      <div className="mt-[12px]">
                        <Alert type="info" showIcon message={$t('平均分')} description={fmt(evaluation.average_score)} />
                        <Table
                          className="mt-[8px]"
                          rowKey="prompt"
                          size="small"
                          pagination={false}
                          dataSource={evaluation.cases || []}
                          columns={[
                            { title: $t('提示词'), dataIndex: 'prompt', key: 'prompt' },
                            { title: $t('评分'), dataIndex: 'score', key: 'score', render: fmt },
                            { title: $t('理由'), dataIndex: 'rationale', key: 'rationale' },
                            { title: $t('错误'), dataIndex: 'error', key: 'error' }
                          ]}
                        />
                      </div>
                    )}
                  </Card>
                </Col>
              </Row>
            )
          },
          {
            key: 'notify',
            label: $t('通知配置'),
            children: (
              <Card title={$t('告警触达通道')} loading={loading}>
                <Alert className="mb-[16px]" type="info" showIcon message={$t('告警哪里去？')} description={$t('配置后，预算告警 / 异常检测的结果会通过所选通道主动送达；后台每 5 分钟评估一次，仅在状态恶化（如 ok→warning→exceeded）时去重发送。')} />
                <Form form={notifyForm} layout="vertical" onFinish={saveNotify} initialValues={{ webhook_method: 'POST', alert_threshold: 0.8 }}>
                  <Descriptions title={$t('通用 Webhook（Slack / 飞书 / 钉钉 / 任意 HTTP）')} column={1} bordered size="small" className="mb-[12px]" />
                  <Form.Item label={$t('启用 Webhook')} name="webhook_enabled" valuePropName="checked">
                    <Input type="checkbox" />
                  </Form.Item>
                  <Form.Item label={$t('Webhook 地址')} name="webhook_url">
                    <Input placeholder="https://webhook.site/xxxx 或 https://open.feishu.cn/..." />
                  </Form.Item>
                  <Form.Item label={$t('请求方法')} name="webhook_method">
                    <Input placeholder="POST" />
                  </Form.Item>
                  <Form.Item label={$t('请求头（每行 Key: Value）')} name="webhook_headers">
                    <Input.TextArea rows={2} placeholder={'Authorization: Bearer xxxx'} />
                  </Form.Item>
                  <Form.Item label={$t('消息模板（Go text/template，可用 {{.Title}} {{.Body}} {{.Severity}} 等）')} name="webhook_template">
                    <Input.TextArea rows={3} placeholder={'留空则发送默认 JSON 结构'} />
                  </Form.Item>

                  <Descriptions title={$t('邮件（SMTP）')} column={1} bordered size="small" className="mb-[12px] mt-[16px]" />
                  <Alert className="mb-[12px]" type="warning" showIcon message={$t('SMTP 需先在 config.yml 配置 smtp: 段')} description={$t('未配置时「邮件」通道在发送时返回明确错误。')} />
                  <Form.Item label={$t('启用邮件')} name="email_enabled" valuePropName="checked">
                    <Input type="checkbox" />
                  </Form.Item>
                  <Form.Item label={$t('收件人（每行一个）')} name="email_to">
                    <Input.TextArea rows={2} placeholder={'ops@example.com'} />
                  </Form.Item>
                  <Form.Item label={$t('主题模板')} name="email_subject">
                    <Input placeholder="{{.Title}}" />
                  </Form.Item>
                  <Form.Item label={$t('正文模板')} name="email_body">
                    <Input.TextArea rows={3} placeholder="{{.Body}}" />
                  </Form.Item>

                  <Space>
                    <Button type="primary" htmlType="submit" loading={loading}>{$t('保存')}</Button>
                    <Button onClick={() => sendTestNotification('webhook')}>{$t('发送 Webhook 测试')}</Button>
                    <Button onClick={() => sendTestNotification('email')}>{$t('发送邮件测试')}</Button>
                    <Button onClick={loadNotify}>{$t('刷新')}</Button>
                  </Space>
                </Form>
              </Card>
            )
          },
          {
            key: 'consumer',
            label: $t('消费者/Key 预算'),
            children: (
              <Card
                title={$t('按消费者（订阅者）预算与配额')}
                loading={loading}
                extra={<Button onClick={seedSample}>{$t('生成示例用量数据')}</Button>}
              >
                <Alert className="mb-[12px]" type="info" showIcon message={$t('按租户/按 Key 的支出上限')} description={$t('为某个消费者（即用量表中的 subscriber，示例数据含 team-alpha / team-beta / internal-app）配置月度支出上限；开启「强制」后会把一条按消费者限流的策略下发到数据面（实际限流由数据面执行）。')} />
                <Form form={consumerForm} layout="vertical" onFinish={saveConsumerBudgets}>
                  <Form.List name="items">
                    {(fields, { add, remove }) => (
                      <>
                        {fields.map((field) => (
                          <Card key={field.key} size="small" className="mb-[12px]" title={$t('消费者')}>
                            <Row gutter={[12, 12]}>
                              <Col xs={24} md={8}>
                                <Form.Item label={$t('消费者 ID')} name={[field.name, 'consumer']} rules={[{ required: true }]}>
                                  <Input placeholder="team-alpha" />
                                </Form.Item>
                              </Col>
                              <Col xs={12} md={4}>
                                <Form.Item label={$t('月度预算')} name={[field.name, 'monthly_budget']}>
                                  <InputNumber className="w-full" min={0} step={100} precision={2} />
                                </Form.Item>
                              </Col>
                              <Col xs={12} md={4}>
                                <Form.Item label={$t('预警阈值')} name={[field.name, 'alert_threshold']}>
                                  <InputNumber className="w-full" min={0} max={1} step={0.05} precision={2} />
                                </Form.Item>
                              </Col>
                              <Col xs={12} md={4}>
                                <Form.Item label={$t('配额(QPS)')} name={[field.name, 'qps']}>
                                  <InputNumber className="w-full" min={1} step={1} precision={0} />
                                </Form.Item>
                              </Col>
                              <Col xs={24} md={4} className="flex items-end">
                                <Space>
                                  <Form.Item name={[field.name, 'enabled']} valuePropName="checked">
                                    <Input type="checkbox" /> {$t('启用')}
                                  </Form.Item>
                                  <Form.Item name={[field.name, 'enforce']} valuePropName="checked">
                                    <Input type="checkbox" /> {$t('强制')}
                                  </Form.Item>
                                </Space>
                              </Col>
                            </Row>
                            <Space>
                              <Button size="small" onClick={() => loadConsumerStatus(consumerForm.getFieldValue(['items', field.name, 'consumer']))}>
                                {$t('查看状态')}
                              </Button>
                              <Button size="small" type="primary" onClick={() => publishConsumerQuota(consumerForm.getFieldValue(['items', field.name, 'consumer']))}>
                                {$t('下发配额策略')}
                              </Button>
                              <Button size="small" type="link" danger onClick={() => remove(field.name)}>{$t('删除')}</Button>
                            </Space>
                            {(() => {
                              const c = consumerForm.getFieldValue(['items', field.name, 'consumer'])
                              const st = c ? consumerStatus[c] : null
                              return st ? (
                                <Descriptions className="mt-[8px]" column={2} bordered size="small">
                                  <Descriptions.Item label={$t('已配置')}>{st.configured ? $t('是') : $t('否')}</Descriptions.Item>
                                  <Descriptions.Item label={$t('本月支出')}>{fmt(st.month_spend)}</Descriptions.Item>
                                  <Descriptions.Item label={$t('占比')}>{(st.ratio * 100).toFixed(0)}%</Descriptions.Item>
                                  <Descriptions.Item label={$t('状态')}>
                                    <Tag color={st.level === 'exceeded' ? 'red' : st.level === 'warning' ? 'orange' : 'green'}>{st.level}</Tag>
                                  </Descriptions.Item>
                                </Descriptions>
                              ) : null
                            })()}
                          </Card>
                        ))}
                        <Button type="dashed" onClick={() => add({ consumer: '', monthly_budget: 0, alert_threshold: 0.8, qps: 10, enabled: false, enforce: false })} block>
                          {$t('添加消费者')}
                        </Button>
                      </>
                    )}
                  </Form.List>
                  <Space className="mt-[12px]">
                    <Button type="primary" htmlType="submit" loading={loading}>{$t('保存')}</Button>
                    <Button onClick={loadConsumerBudgets}>{$t('刷新')}</Button>
                  </Space>
                </Form>
              </Card>
            )
          }
        ]}
      />
    </InsidePage>
  )
}

interface ProviderCost {
  per_call: number
  per_1k_input_token: number
  per_1k_output_token: number
}

interface ConsumerBudgetRow {
  consumer: string
  monthly_budget: number
  alert_threshold: number
  qps: number
  enabled: boolean
  enforce: boolean
}
