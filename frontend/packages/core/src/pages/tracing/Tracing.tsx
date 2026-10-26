import InsidePage from '@common/components/aoplatform/InsidePage'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  DatePicker,
  Descriptions,
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

export default function Tracing() {
  const { fetchData } = useGlobalContext()
  const { message: msgApi } = App.useApp()
  const [enabled, setEnabled] = useState(true)
  const [loading, setLoading] = useState(false)

  const [range, setRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([dayjs().subtract(30, 'day'), dayjs()])
  const rangeRef = useRef(range)
  rangeRef.current = range

  const [overview, setOverview] = useState<any>(null)
  const [list, setList] = useState<any>(null)

  const loadEnabled = () => {
    fetchData<BasicResponse<{ modules: { modules: ModuleStatus[] } }>>('license/modules', { method: 'GET' })
      .then((resp) => {
        if (resp?.code === STATUS_CODE.SUCCESS) {
          const m = (resp.data.modules?.modules || []).find((x) => x.name === 'tracing')
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

  const windowParams = () => {
    const [s, e] = rangeRef.current
    return { start: Math.floor(s.toDate().getTime() / 1000), end: Math.floor(e.toDate().getTime() / 1000) }
  }

  const loadOverview = () => {
    setLoading(true)
    api<{ overview: any }>('tracing/overview', { method: 'GET', eoParams: windowParams() })
      .then((d) => setOverview(d.overview))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  const seedSample = () => {
    setLoading(true)
    api<{ result: { status: string; days: number } }>('tracing/seed', { method: 'POST', eoParams: { days: 30 } })
      .then(() => {
        msgApi.success($t('已生成示例链路追踪数据'))
        loadOverview()
        loadList(1)
      })
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  const loadList = (page: number) => {
    setLoading(true)
    api<{ list: any }>('tracing/list', { method: 'GET', eoParams: { ...windowParams(), page, page_size: 20 } })
      .then((d) => setList(d.list))
      .catch((e) => msgApi.error(e.message))
      .finally(() => setLoading(false))
  }

  if (enabled && !overview && !list) {
    loadEnabled()
    loadOverview()
    loadList(1)
  }

  const locked = (
    <Alert
      className="mb-[16px]"
      type="warning"
      showIcon
      message={$t('该能力属于企业版（链路追踪与 LLM 可观测性）')}
      description={$t('当前 License 未包含 tracing 模块，无法使用链路追踪。请升级企业版后使用。')}
    />
  )

  const truncate = (s: string, n = 60) => (s && s.length > n ? `${s.slice(0, n)}…` : s)

  return (
    <InsidePage
      pageTitle={$t('链路追踪')}
      showBorder={false}
      contentClassName="pr-PAGE_INSIDE_X"
      scrollPage={false}
      description={$t('逐次 LLM 调用的完整 prompt/response、延迟拆解与 token 明细——排障与细粒度成本归因的可观测性下一层')}
    >
      {!enabled && locked}

      <Tabs
        defaultActiveKey="overview"
        onChange={(key) => {
          if (key === 'overview') loadOverview()
          if (key === 'list') loadList(1)
        }}
        items={[
          {
            key: 'overview',
            label: $t('概览'),
            children: (
              <Card
                title={$t('LLM 调用概览')}
                loading={loading}
                extra={
                  <Space>
                    <DatePicker.RangePicker value={range} onChange={(v) => v && setRange(v as [dayjs.Dayjs, dayjs.Dayjs])} />
                    <Button onClick={seedSample}>{$t('生成示例追踪数据')}</Button>
                    <Button type="primary" onClick={loadOverview}>{$t('刷新')}</Button>
                  </Space>
                }
              >
                {overview && (
                  <>
                    <Descriptions column={3} bordered size="small" className="mb-[16px]">
                      <Descriptions.Item label={$t('总调用')}>{fmtInt(overview.total_calls)}</Descriptions.Item>
                      <Descriptions.Item label={$t('输入Token')}>{fmtInt(overview.input_tokens)}</Descriptions.Item>
                      <Descriptions.Item label={$t('输出Token')}>{fmtInt(overview.output_tokens)}</Descriptions.Item>
                      <Descriptions.Item label={$t('总Token')}>{fmtInt(overview.total_tokens)}</Descriptions.Item>
                      <Descriptions.Item label={$t('时间范围起')}>{dayjs.unix(overview.window?.start || 0).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>
                      <Descriptions.Item label={$t('时间范围止')}>{dayjs.unix(overview.window?.end || 0).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>
                    </Descriptions>
                    <Table
                      rowKey="provider"
                      size="small"
                      pagination={false}
                      dataSource={overview.providers || []}
                      columns={[
                        { title: $t('供应商'), dataIndex: 'provider', key: 'provider' },
                        { title: $t('调用数'), dataIndex: 'calls', key: 'calls', render: fmtInt }
                      ]}
                    />
                  </>
                )}
                {!overview && <span className="text-[13px] text-[#8a8aa0]">{$t('选择时间范围并点击「刷新」；或先生成示例数据')}</span>}
              </Card>
            )
          },
          {
            key: 'list',
            label: $t('调用明细'),
            children: (
              <Card
                title={$t('逐次调用记录')}
                loading={loading}
                extra={
                  <Space>
                    <DatePicker.RangePicker value={range} onChange={(v) => v && setRange(v as [dayjs.Dayjs, dayjs.Dayjs])} />
                    <Button type="primary" onClick={() => loadList(1)}>{$t('查询')}</Button>
                  </Space>
                }
              >
                {list && (
                  <Table
                    rowKey="id"
                    size="small"
                    dataSource={list.items || []}
                    pagination={{
                      current: list.page,
                      pageSize: list.page_size,
                      total: list.total,
                      onChange: (p) => loadList(p)
                    }}
                    columns={[
                      { title: $t('时间'), dataIndex: 'created_at', key: 'created_at' },
                      { title: $t('服务'), dataIndex: 'service_id', key: 'service_id' },
                      { title: $t('模型'), dataIndex: 'model', key: 'model' },
                      { title: $t('供应商'), dataIndex: 'provider', key: 'provider' },
                      { title: $t('消费者'), dataIndex: 'subscriber', key: 'subscriber' },
                      { title: $t('Token'), key: 'tokens', render: (_, r: any) => fmtInt(r.total_tokens) },
                      { title: $t('延迟(ms)'), dataIndex: 'latency_total_ms', key: 'latency_total_ms', render: fmtInt },
                      { title: $t('状态'), dataIndex: 'status_code', key: 'status_code', render: (c: number) => <Tag color={c >= 500 ? 'red' : c >= 400 ? 'orange' : 'green'}>{c}</Tag> },
                      { title: $t('Prompt'), dataIndex: 'prompt', key: 'prompt', render: (s: string) => truncate(s) },
                      { title: $t('响应'), dataIndex: 'completion', key: 'completion', render: (s: string) => truncate(s) }
                    ]}
                    expandable={{
                      expandedRowRender: (r: any) => (
                        <div className="text-[13px]">
                          <div className="mb-[6px]"><b>{$t('Prompt')}：</b>{r.prompt}</div>
                          <div><b>{$t('Response')}：</b>{r.completion}</div>
                        </div>
                      )
                    }}
                  />
                )}
                {!list && <span className="text-[13px] text-[#8a8aa0]">{$t('点击「查询」加载调用明细')}</span>}
              </Card>
            )
          }
        ]}
      />
    </InsidePage>
  )
}
