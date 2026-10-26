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
import { useEffect, useRef, useState } from 'react'
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

interface TenantRow {
  id: string
  name: string
  monthly_budget: number
  quota_calls: number
  enabled: boolean
}

const fmt = (v: number) => (v == null ? '-' : v.toFixed(2))
const fmtInt = (v: number) => (v == null ? '-' : `${v}`)

export default function MultiTenant() {
  const { fetchData } = useGlobalContext()
  const { message: msgApi } = App.useApp()
  const [enabled, setEnabled] = useState(true)
  const [loading, setLoading] = useState(false)

  const [tenantForm] = Form.useForm<{ items: TenantRow[] }>()
  const [range, setRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([dayjs().subtract(30, 'day'), dayjs()])
  const rangeRef = useRef(range)
  rangeRef.current = range

  const [billing, setBilling] = useState<any>(null)

  const loadEnabled = () => {
    fetchData<BasicResponse<{ modules: { modules: ModuleStatus[] } }>>('license/modules', { method: 'GET' })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          const m = (resp.data.modules?.modules || []).find((x) => x.name === 'multi_tenant')
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

  // ---- tenants ----
  const loadTenants = () => {
    setLoading(true)
    api<{ tenants: { items: { [k: string]: any } } }>('multitenant/tenants', { method: 'GET' })
      .then((d) => {
        const items = Object.keys(d.tenants?.items || {}).map((id) => ({ id, ...d.tenants.items[id] }))
        tenantForm.setFieldsValue({
          items: items.length ? items : [{ id: '', name: '', monthly_budget: 0, quota_calls: 0, enabled: false }]
        })
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }
  const saveTenants = (v: any) => {
    const items: { [k: string]: any } = {}
    ;(v.items || []).forEach((it: any) => {
      const id = (it.id || '').trim()
      if (id) items[id] = { name: it.name || '', monthly_budget: it.monthly_budget || 0, quota_calls: it.quota_calls || 0, enabled: !!it.enabled }
    })
    setLoading(true)
    api<{ tenants: any }>('multitenant/tenants', { method: 'PUT', eoBody: { items } })
      .then(() => {
        msgApi.success($t('租户配置已保存'))
        loadTenants()
        loadBilling()
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  // ---- billing ----
  const windowParams = () => {
    const [s, e] = rangeRef.current
    return { start: Math.floor(s.toDate().getTime() / 1000), end: Math.floor(e.toDate().getTime() / 1000) }
  }
  const loadBilling = () => {
    setLoading(true)
    api<{ billing: any[] }>('multitenant/billing', { method: 'GET', eoParams: windowParams() })
      .then((d) => setBilling(d.billing))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (enabled) {
      loadEnabled()
      loadTenants()
      loadBilling()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const locked = (
    <Alert
      className="mb-[16px]"
      type="warning"
      showIcon
      message={$t('该能力属于企业版（多租户隔离与计量分离）')}
      description={$t('当前 License 未包含 multi_tenant 模块，无法使用多租户隔离与计量。请升级企业版后使用。')}
    />
  )

  return (
    <InsidePage
      pageTitle={$t('多租户隔离与计量分离')}
      showBorder={false}
      contentClassName="pr-PAGE_INSIDE_X"
      scrollPage={false}
      description={$t('按租户（subscriber）做计量分离与计费/配额视图——托管云 SaaS 按量模式的前提')}
    >
      {!enabled && locked}

      <Tabs
        defaultActiveKey="manage"
        onChange={(key) => {
          if (key === 'manage') loadTenants()
          if (key === 'billing') loadBilling()
        }}
        items={[
          {
            key: 'manage',
            label: $t('租户管理'),
            children: (
              <Card title={$t('租户定义（Id 即用量表的 subscriber）')} loading={loading}>
                <Alert className="mb-[12px]" type="info" showIcon message={$t('租户 = 计量订阅者')} description={$t('每个租户的 Id 对应用量表中的 subscriber（示例数据含 team-alpha / team-beta / internal-app）。配置月度预算后，可在「租户计量」中查看各租户支出与预算状态。')} />
                <Form form={tenantForm} layout="vertical" onFinish={saveTenants}>
                  <Form.List name="items">
                    {(fields, { add, remove }) => (
                      <>
                        {fields.map((field) => (
                          <Card key={field.key} size="small" className="mb-[12px]" title={$t('租户')}>
                            <Row gutter={[12, 12]}>
                              <Col xs={24} md={6}>
                                <Form.Item label={$t('租户 ID')} name={[field.name, 'id']} rules={[{ required: true }]}>
                                  <Input placeholder="team-alpha" />
                                </Form.Item>
                              </Col>
                              <Col xs={24} md={6}>
                                <Form.Item label={$t('名称')} name={[field.name, 'name']}>
                                  <Input placeholder={$t('Alpha 团队')} />
                                </Form.Item>
                              </Col>
                              <Col xs={12} md={5}>
                                <Form.Item label={$t('月度预算')} name={[field.name, 'monthly_budget']}>
                                  <InputNumber className="w-full" min={0} step={100} precision={2} />
                                </Form.Item>
                              </Col>
                              <Col xs={12} md={5}>
                                <Form.Item label={$t('调用配额')} name={[field.name, 'quota_calls']}>
                                  <InputNumber className="w-full" min={0} step={1000} precision={0} />
                                </Form.Item>
                              </Col>
                              <Col xs={24} md={2} className="flex items-end">
                                <Form.Item name={[field.name, 'enabled']} valuePropName="checked">
                                  <Input type="checkbox" /> {$t('启用')}
                                </Form.Item>
                              </Col>
                            </Row>
                            <Button size="small" type="link" danger onClick={() => remove(field.name)}>{$t('删除')}</Button>
                          </Card>
                        ))}
                        <Button type="dashed" onClick={() => add({ id: '', name: '', monthly_budget: 0, quota_calls: 0, enabled: false })} block>
                          {$t('添加租户')}
                        </Button>
                      </>
                    )}
                  </Form.List>
                  <Space className="mt-[12px]">
                    <Button type="primary" htmlType="submit" loading={loading}>{$t('保存')}</Button>
                    <Button onClick={loadTenants}>{$t('刷新')}</Button>
                  </Space>
                </Form>
              </Card>
            )
          },
          {
            key: 'billing',
            label: $t('租户计量'),
            children: (
              <Card
                title={$t('各租户支出与预算')}
                loading={loading}
                extra={
                  <Space>
                    <DatePicker.RangePicker value={range} onChange={(v) => v && setRange(v as [dayjs.Dayjs, dayjs.Dayjs])} />
                    <Button type="primary" onClick={loadBilling}>{$t('查询')}</Button>
                  </Space>
                }
              >
                {billing && (
                  <Table
                    rowKey="id"
                    size="small"
                    pagination={false}
                    dataSource={billing}
                    columns={[
                      { title: $t('租户'), dataIndex: 'name', key: 'name', render: (n: string, r: any) => `${n || r.id} (${r.id})` },
                      { title: $t('调用数'), dataIndex: 'calls', key: 'calls', render: fmtInt },
                      { title: $t('Token'), dataIndex: 'total_tokens', key: 'total_tokens', render: fmtInt },
                      { title: $t('支出'), dataIndex: 'bill', key: 'bill', render: fmt },
                      { title: $t('月度预算'), dataIndex: 'monthly_budget', key: 'monthly_budget', render: fmt },
                      { title: $t('占比'), dataIndex: 'ratio', key: 'ratio', render: (v: number) => `${((v || 0) * 100).toFixed(0)}%` },
                      { title: $t('状态'), dataIndex: 'level', key: 'level', render: (l: string) => <Tag color={l === 'exceeded' ? 'red' : l === 'warning' ? 'orange' : 'green'}>{l || 'ok'}</Tag> }
                    ]}
                  />
                )}
                {!billing && <span className="text-[13px] text-[#8a8aa0]">{$t('点击「查询」查看各租户计量；可先到「计费与返佣」配置单价，或生成示例用量数据。')}</span>}
              </Card>
            )
          }
        ]}
      />
    </InsidePage>
  )
}
