import InsidePage from '@common/components/aoplatform/InsidePage'
import PageList, { PageProColumns } from '@common/components/aoplatform/PageList'
import { App, Alert, Button, DatePicker, Input, Select, Space } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { useGlobalContext } from '@common/contexts/GlobalStateContext'
import { useFetch } from '@common/hooks/http'
import { $t } from '@common/locales'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { ActionType } from '@ant-design/pro-components'
import type { Dayjs } from 'dayjs'

type AuditRecord = {
  id: number
  operator: string
  resourceType: string
  resourceId: string
  action: string
  detail: string
  before: string
  after: string
  ip: string
  createAt: string
}

const RESOURCE_TYPES = [
  { label: $t('策略'), value: 'strategy' },
  { label: $t('密钥'), value: 'ai-key' },
  { label: $t('集群'), value: 'cluster' },
  { label: $t('角色'), value: 'role' },
  { label: $t('成员'), value: 'member' },
  { label: $t('服务'), value: 'service' }
]

const ACTIONS = [
  { label: $t('创建'), value: 'create' },
  { label: $t('更新'), value: 'update' },
  { label: $t('删除'), value: 'delete' },
  { label: $t('发布'), value: 'publish' },
  { label: $t('启用'), value: 'enable' },
  { label: $t('停用'), value: 'disable' },
  { label: $t('登录'), value: 'login' }
]

const AuditList = () => {
  const { message } = App.useApp()
  const { fetchData } = useFetch()
  const { state } = useGlobalContext()
  const pageListRef = useRef<ActionType>(null)

  // 企业版审计模块授权状态
  const [auditEnabled, setAuditEnabled] = useState<boolean>(true)

  // 过滤器
  const [keyword, setKeyword] = useState<string>('')
  const [resourceType, setResourceType] = useState<string>('')
  const [action, setAction] = useState<string>('')
  const [operator, setOperator] = useState<string>('')
  const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null)

  useEffect(() => {
    fetchData<BasicResponse<{ modules: { modules: { name: string; enabled: boolean }[] } }>>('license/modules', { method: 'GET' })
      .then((response) => {
        const { code, data } = response
        if (code === STATUS_CODE.SUCCESS) {
          const mod = (data.modules.modules || []).find((m: any) => m.name === 'audit')
          setAuditEnabled(!!(mod && mod.enabled))
        }
      })
      .catch(() => { /* 忽略，默认放行 */ })
  }, [])

  const buildParams = () => {
    const params: Record<string, any> = {
      keyword,
      resourceType,
      action,
      operator,
      startTime: 0,
      endTime: 0
    }
    if (dateRange && dateRange[0] && dateRange[1]) {
      params.startTime = dateRange[0].unix()
      params.endTime = dateRange[1].unix()
    }
    return params
  }

  const getAuditList = (params: any & { pageSize: number; current: number }) => {
    const query = buildParams()
    return fetchData<BasicResponse<{ list: AuditRecord[]; total: number }>>('audit/logs', {
      method: 'GET',
      eoParams: {
        keyword: query.keyword,
        resourceType: query.resourceType,
        action: query.action,
        operator: query.operator,
        startTime: query.startTime,
        endTime: query.endTime,
        pageNum: params.current,
        pageSize: params.pageSize
      },
      eoTransformKeys: ['resource_type', 'resource_id', 'before_value', 'after_value', 'create_at']
    }).then((response) => {
      const { code, data, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        return { data: data.audit?.list || [], total: data.audit?.total || 0, success: true }
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
        return { data: [], success: false }
      }
    }).catch(() => ({ data: [], success: false }))
  }

  const onExport = () => {
    const query = buildParams()
    message.loading($t(RESPONSE_TIPS.loading))
    fetchData<BasicResponse<{ file_name: string; content: string }>>('audit/report/export', {
      method: 'GET',
      eoParams: {
        keyword: query.keyword,
        resourceType: query.resourceType,
        action: query.action,
        operator: query.operator,
        startTime: query.startTime,
        endTime: query.endTime
      }
    }).then((response) => {
      message.destroy()
      const { code, data, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        const blob = new Blob([data.audit.content], { type: 'text/csv;charset=utf-8' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = data.audit.file_name || 'audit-report.csv'
        a.click()
        URL.revokeObjectURL(url)
        message.success($t(RESPONSE_TIPS.success))
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
      }
    }).catch(() => message.destroy())
  }

  const columns: PageProColumns<AuditRecord>[] = [
    { title: $t('操作时间'), dataIndex: 'createAt', width: 170 },
    { title: $t('操作人'), dataIndex: 'operator', width: 120 },
    { title: $t('来源IP'), dataIndex: 'ip', width: 130 },
    { title: $t('资源类型'), dataIndex: 'resourceType', width: 110 },
    { title: $t('资源ID'), dataIndex: 'resourceId', ellipsis: true },
    { title: $t('动作'), dataIndex: 'action', width: 90 },
    { title: $t('操作描述'), dataIndex: 'detail', ellipsis: true }
  ]

  const reload = () => pageListRef.current?.reload()

  return (
    <InsidePage pageTitle={$t('审计日志')} showBorder={false} contentClassName="pr-PAGE_INSIDE_X" scrollPage={false} description={$t('记录网关关键操作，支撑合规审计与追溯')}>
      {!auditEnabled && (
        <Alert
          className="mb-[12px]"
          type="warning"
          showIcon
          message={$t('该能力属于企业版（审计与合规报表）')}
          description={$t('当前 License 未包含 audit 模块，无法查看审计日志或导出报表。请升级企业版后使用。')}
        />
      )}
      <PageList<AuditRecord>
        id="audit_logs_list"
        ref={pageListRef}
        columns={columns}
        request={async (params: any & { pageSize: number; current: number }) => getAuditList(params)}
        search={false}
        scroll={{ x: 1000 }}
        toolBarRender={() => [
          <Space key="filters" wrap>
            <Input
              className="w-INPUT_NORMAL"
              placeholder={$t('关键字（操作人/资源ID/描述）')}
              allowClear
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onPressEnter={reload}
            />
            <Select
              className="w-INPUT_NORMAL"
              placeholder={$t('资源类型')}
              allowClear
              value={resourceType || undefined}
              options={RESOURCE_TYPES}
              onChange={(v) => setResourceType(v || '')}
            />
            <Select
              className="w-INPUT_NORMAL"
              placeholder={$t('动作')}
              allowClear
              value={action || undefined}
              options={ACTIONS}
              onChange={(v) => setAction(v || '')}
            />
            <Input
              className="w-INPUT_NORMAL"
              placeholder={$t('操作人')}
              allowClear
              value={operator}
              onChange={(e) => setOperator(e.target.value)}
              onPressEnter={reload}
            />
            <DatePicker.RangePicker
              value={dateRange as any}
              onChange={(v) => setDateRange(v as any)}
            />
            <Button type="primary" onClick={reload}>{$t('查询')}</Button>
          </Space>
        ]}
        afterNewBtn={[
          <Button key="export" disabled={!auditEnabled} onClick={onExport}>{$t('导出报表')}</Button>
        ]}
      />
    </InsidePage>
  )
}

export default AuditList
