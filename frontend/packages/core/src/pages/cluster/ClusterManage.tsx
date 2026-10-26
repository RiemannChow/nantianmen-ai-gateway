import InsidePage from '@common/components/aoplatform/InsidePage'
import PageList, { PageProColumns } from '@common/components/aoplatform/PageList'
import { App, Alert, Button, Form, Input, InputNumber, Modal, Select, Space, Popconfirm, Tag } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { useGlobalContext } from '@common/contexts/GlobalStateContext'
import { useFetch } from '@common/hooks/http'
import { $t } from '@common/locales'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { ActionType } from '@ant-design/pro-components'

type ClusterRecord = {
  uuid: string
  name: string
  resume: string
  role: string
  priority: number
  enabled: boolean
  healthy: number
  node_count: number
  create_time: string
  update_time: string
}

type DRView = {
  total: number
  enabled: number
  healthy: number
  primary: number
  backup: number
  clusters: ClusterRecord[]
}

const ROLE_OPTIONS = [
  { label: $t('主集群'), value: 'primary' },
  { label: $t('备集群'), value: 'backup' },
  { label: $t('独立集群'), value: 'standalone' }
]

const ClusterManage = () => {
  const { message } = App.useApp()
  const { fetchData } = useFetch()
  const pageListRef = useRef<ActionType>(null)
  const [form] = Form.useForm()

  // 企业版多集群容灾模块授权状态
  const [drEnabled, setDrEnabled] = useState<boolean>(true)
  const [modalOpen, setModalOpen] = useState<boolean>(false)
  const [editing, setEditing] = useState<ClusterRecord | null>(null)
  const [dr, setDr] = useState<DRView | null>(null)

  useEffect(() => {
    fetchData<BasicResponse<{ modules: { modules: { name: string; enabled: boolean }[] } }>>('license/modules', { method: 'GET' })
      .then((response) => {
        const { code, data } = response
        if (code === STATUS_CODE.SUCCESS) {
          const mod = (data.modules.modules || []).find((m: any) => m.name === 'multi_cluster')
          setDrEnabled(!!(mod && mod.enabled))
        }
      })
      .catch(() => {
        /* 忽略，默认放行 */
      })
  }, [])

  const loadDr = () => {
    fetchData<BasicResponse<{ dr: DRView }>>('cluster/dr', { method: 'GET' })
      .then((response) => {
        const { code, data } = response
        if (code === STATUS_CODE.SUCCESS) {
          setDr(data.dr)
        }
      })
      .catch(() => {
        /* 忽略 */
      })
  }

  useEffect(() => {
    if (drEnabled) loadDr()
  }, [drEnabled])

  const getClusterList = () => {
    return fetchData<BasicResponse<{ clusters: ClusterRecord[] }>>('cluster/list', { method: 'GET' })
      .then((response) => {
        const { code, data, msg } = response
        if (code === STATUS_CODE.SUCCESS) {
          const list = data.clusters || []
          return { data: list, total: list.length, success: true }
        } else {
          message.error(msg || $t(RESPONSE_TIPS.error))
          return { data: [], success: false }
        }
      })
      .catch(() => ({ data: [], success: false }))
  }

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ role: 'standalone', priority: 100, enabled: true })
    setModalOpen(true)
  }

  const openEdit = (record: ClusterRecord) => {
    setEditing(record)
    form.setFieldsValue({
      name: record.name,
      resume: record.resume,
      address: '',
      role: record.role,
      priority: record.priority,
      enabled: record.enabled
    })
    setModalOpen(true)
  }

  const onModalOk = () => {
    form.validateFields().then((values) => {
      if (editing) {
        const body: Record<string, any> = {}
        if (values.name !== undefined && values.name !== '') body.name = values.name
        if (values.resume !== undefined) body.resume = values.resume
        if (values.role !== undefined && values.role !== '') body.role = values.role
        if (values.priority !== undefined) body.priority = values.priority
        if (values.enabled !== undefined) body.enabled = values.enabled
        return fetchData<BasicResponse<{ cluster: ClusterRecord }>>(`cluster/${editing.uuid}`, {
          method: 'PUT',
          eoParams: { id: editing.uuid },
          eoBody: body
        })
          .then((response) => {
            const { code, msg } = response
            if (code === STATUS_CODE.SUCCESS) {
              message.success(msg || $t(RESPONSE_TIPS.success))
              setModalOpen(false)
              pageListRef.current?.reload()
              loadDr()
              return Promise.resolve(true)
            } else {
              message.error(msg || $t(RESPONSE_TIPS.error))
              return Promise.reject(msg || $t(RESPONSE_TIPS.error))
            }
          })
          .catch((e) => Promise.reject(e))
      }
      return fetchData<BasicResponse<{ cluster: ClusterRecord }>>('cluster', {
        method: 'POST',
        eoBody: {
          name: values.name,
          resume: values.resume,
          address: values.address,
          role: values.role,
          priority: values.priority
        }
      })
        .then((response) => {
          const { code, msg } = response
          if (code === STATUS_CODE.SUCCESS) {
            message.success(msg || $t(RESPONSE_TIPS.success))
            setModalOpen(false)
            pageListRef.current?.reload()
            loadDr()
            return Promise.resolve(true)
          } else {
            message.error(msg || $t(RESPONSE_TIPS.error))
            return Promise.reject(msg || $t(RESPONSE_TIPS.error))
          }
        })
        .catch((e) => Promise.reject(e))
    })
  }

  const onDelete = (record: ClusterRecord) => {
    fetchData<BasicResponse<{ id: string }>>(`cluster/${record.uuid}`, {
      method: 'DELETE',
      eoParams: { id: record.uuid }
    })
      .then((response) => {
        const { code, msg } = response
        if (code === STATUS_CODE.SUCCESS) {
          message.success(msg || $t(RESPONSE_TIPS.success))
          pageListRef.current?.reload()
          loadDr()
        } else {
          message.error(msg || $t(RESPONSE_TIPS.error))
        }
      })
      .catch(() => {
        /* 忽略 */
      })
  }

  const onProbe = (record: ClusterRecord) => {
    message.loading($t(RESPONSE_TIPS.loading))
    fetchData<BasicResponse<{ nodes: any[] }>>(`cluster/${record.uuid}/probe`, {
      method: 'POST',
      eoParams: { id: record.uuid }
    })
      .then((response) => {
        const { code, msg } = response
        message.destroy()
        if (code === STATUS_CODE.SUCCESS) {
          message.success($t('探测成功，集群健康状态已更新'))
          pageListRef.current?.reload()
          loadDr()
        } else {
          message.error(msg || $t(RESPONSE_TIPS.error))
        }
      })
      .catch(() => message.destroy())
  }

  const roleTag = (role: string) => {
    switch (role) {
      case 'primary':
        return <Tag color="blue">{$t('主集群')}</Tag>
      case 'backup':
        return <Tag color="green">{$t('备集群')}</Tag>
      default:
        return <Tag>{$t('独立集群')}</Tag>
    }
  }

  const columns: PageProColumns<ClusterRecord>[] = [
    { title: $t('集群名称'), dataIndex: 'name', width: 160 },
    { title: $t('描述'), dataIndex: 'resume', ellipsis: true },
    {
      title: $t('角色'),
      dataIndex: 'role',
      width: 110,
      render: (_, record) => roleTag(record.role)
    },
    { title: $t('优先级'), dataIndex: 'priority', width: 90 },
    {
      title: $t('容灾发布'),
      dataIndex: 'enabled',
      width: 100,
      render: (_, record) => (record.enabled ? <Tag color="blue">{$t('启用')}</Tag> : <Tag color="default">{$t('停用')}</Tag>)
    },
    {
      title: $t('健康状态'),
      dataIndex: 'healthy',
      width: 100,
      render: (_, record) =>
        record.healthy === 1 ? <Tag color="green">{$t('健康')}</Tag> : <Tag color="red">{$t('异常')}</Tag>
    },
    { title: $t('节点数'), dataIndex: 'node_count', width: 80 },
    {
      title: $t('操作'),
      key: 'option',
      width: 220,
      btnNums: 3,
      render: (_, record) => (
        <Space>
          <Button type="link" disabled={!drEnabled} onClick={() => onProbe(record)}>
            {$t('探测')}
          </Button>
          <Button type="link" disabled={!drEnabled} onClick={() => openEdit(record)}>
            {$t('编辑')}
          </Button>
          <Popconfirm title={$t('确认删除该集群？')} onConfirm={() => onDelete(record)} disabled={!drEnabled}>
            <Button type="link" danger disabled={!drEnabled}>
              {$t('删除')}
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ]

  const reload = () => pageListRef.current?.reload()

  return (
    <InsidePage
      pageTitle={$t('集群管理')}
      showBorder={false}
      contentClassName="pr-PAGE_INSIDE_X"
      scrollPage={false}
      description={$t('接入与管理 Apinto 数据面集群，并配置多集群容灾（主备角色、优先级与健康探测）')}
    >
      {!drEnabled && (
        <Alert
          className="mb-[12px]"
          type="warning"
          showIcon
          message={$t('该能力属于企业版（多集群容灾）')}
          description={$t('当前 License 未包含 multi_cluster 模块，无法管理多集群或配置容灾。请升级企业版后使用。')}
        />
      )}

      {drEnabled && dr && (
        <div className="mb-[16px] flex gap-[12px] flex-wrap">
          <StatCard label={$t('集群总数')} value={dr.total} />
          <StatCard label={$t('容灾启用')} value={dr.enabled} />
          <StatCard label={$t('健康集群')} value={dr.healthy} />
          <StatCard label={$t('主集群')} value={dr.primary} color="#1677ff" />
          <StatCard label={$t('备集群')} value={dr.backup} color="#52c41a" />
        </div>
      )}

      <PageList<ClusterRecord>
        id="cluster_manage_list"
        ref={pageListRef}
        columns={columns}
        request={async () => getClusterList()}
        search={false}
        showPagination={false}
        scroll={{ x: 1100 }}
        addNewBtnTitle={$t('新增集群')}
        addNewBtnDisabled={!drEnabled}
        onAddNewBtnClick={openCreate}
      />

      <Modal
        title={editing ? $t('编辑集群') : $t('新增集群')}
        open={modalOpen}
        onOk={onModalOk}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
        okText={$t('保存')}
        cancelText={$t('取消')}
      >
        <Form form={form} layout="vertical" name="clusterForm">
          <Form.Item label={$t('集群名称')} name="name" rules={[{ required: true, whitespace: true, message: $t('请输入集群名称') }]}>
            <Input className="w-INPUT_NORMAL" placeholder={$t('请输入集群名称')} />
          </Form.Item>
          <Form.Item label={$t('描述')} name="resume">
            <Input className="w-INPUT_NORMAL" placeholder={$t('请输入描述')} />
          </Form.Item>
          {!editing && (
            <Form.Item
              label={$t('数据面管理地址')}
              name="address"
              extra={$t('Apinto 集群管理地址，例如 http://127.0.0.1:9401')}
              rules={[{ required: true, whitespace: true, message: $t('请输入数据面管理地址') }]}
            >
              <Input className="w-INPUT_NORMAL" placeholder="http://host:9401" />
            </Form.Item>
          )}
          <Form.Item label={$t('容灾角色')} name="role" rules={[{ required: true, message: $t('请选择容灾角色') }]}>
            <Select className="w-INPUT_NORMAL" options={ROLE_OPTIONS} />
          </Form.Item>
          <Form.Item label={$t('优先级')} name="priority" extra={$t('数值越小越优先，发布时主集群优先，主不可用时自动切到备集群')}>
            <InputNumber className="w-INPUT_NORMAL" min={0} max={999} />
          </Form.Item>
          <Form.Item label={$t('容灾发布')} name="enabled" valuePropName="checked">
            <Select
              className="w-INPUT_NORMAL"
              options={[
                { label: $t('启用'), value: true },
                { label: $t('停用'), value: false }
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </InsidePage>
  )
}

const StatCard = ({ label, value, color }: { label: string; value: number; color?: string }) => (
  <div className="rounded-md border border-[#e5e7eb] px-[16px] py-[10px] min-w-[120px]">
    <div className="text-[12px] text-[#6b7280]">{label}</div>
    <div className="text-[22px] font-semibold" style={color ? { color } : undefined}>
      {value}
    </div>
  </div>
)

export default ClusterManage
