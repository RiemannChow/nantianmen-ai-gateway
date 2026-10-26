import { ActionType } from '@ant-design/pro-components'
import InsidePage from '@common/components/aoplatform/InsidePage'
import PageList, { PageProColumns } from '@common/components/aoplatform/PageList'
import WithPermission from '@common/components/aoplatform/WithPermission'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { useFetch } from '@common/hooks/http'
import { $t } from '@common/locales/index.ts'
import { App, Button, Tag, Typography } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { PromptTemplate, PromptManagementHandle } from './type'
import TableBtnWithPermission from '@common/components/aoplatform/TableBtnWithPermission'
import PromptEditorModal from './PromptEditorModal'

const PromptManagement = () => {
  const pageListRef = useRef<ActionType>(null)
  const [searchWord, setSearchWord] = useState<string>('')
  const { modal, message } = App.useApp()
  const [editing, setEditing] = useState<PromptTemplate | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const { fetchData } = useFetch()

  const requestTemplates = () => {
    return fetchData<BasicResponse<{ list: PromptTemplate[] }>>('ai/prompts', {
      method: 'GET',
      eoParams: { keyword: searchWord },
      eoTransformKeys: ['updated_at']
    })
      .then((response) => {
        const { code, data } = response
        if (code === STATUS_CODE.SUCCESS) {
          return { data: data.list, total: data.list.length, success: true }
        }
        message.error(response.msg || $t(RESPONSE_TIPS.error))
        return { data: [], success: false }
      })
      .catch(() => ({ data: [], success: false }))
  }

  const openCreate = () => {
    setEditing(null)
    setModalOpen(true)
  }

  const openEdit = (record: PromptTemplate) => {
    setEditing(record)
    setModalOpen(true)
  }

  const handleDelete = (id: string) => {
    modal.confirm({
      title: $t('删除提示词'),
      content: $t('确定要删除该提示词模板吗？'),
      onOk: () => {
        return fetchData<BasicResponse<null>>('ai/prompt', {
          method: 'DELETE',
          eoParams: { id }
        })
          .then((response) => {
            if (response.code === STATUS_CODE.SUCCESS) {
              message.success($t('删除成功'))
              pageListRef.current?.reload()
            } else {
              message.error(response.msg || $t(RESPONSE_TIPS.error))
            }
          })
          .catch(() => message.error($t(RESPONSE_TIPS.error)))
      }
    })
  }

  const formatTime = (ts: string) => {
    const n = Number(ts)
    if (!n) return '-'
    return new Date(n * 1000).toLocaleString()
  }

  const columns: PageProColumns<PromptTemplate>[] = [
    {
      title: $t('名称'),
      dataIndex: 'name',
      ellipsis: true,
      key: 'name'
    },
    {
      title: $t('标签'),
      dataIndex: 'tags',
      ellipsis: true,
      key: 'tags',
      render: (dom: React.ReactNode, record: PromptTemplate) => (
        <span>
          {(record.tags || []).map((t) => (
            <Tag key={t}>{t}</Tag>
          ))}
        </span>
      )
    },
    {
      title: $t('内容'),
      dataIndex: 'content',
      ellipsis: true,
      key: 'content',
      render: (dom: React.ReactNode, record: PromptTemplate) => (
        <Typography.Text>{record.content?.replace(/<[^>]+>/g, '') || '-'}</Typography.Text>
      )
    },
    {
      title: $t('更新时间'),
      dataIndex: 'updatedAt',
      width: 180,
      ellipsis: true,
      key: 'updatedAt',
      render: (dom: React.ReactNode, record: PromptTemplate) => (
        <span>{formatTime(record.updatedAt)}</span>
      )
    },
    {
      title: '',
      key: 'option',
      btnNums: 2,
      width: 90,
      fixed: 'right',
      valueType: 'option',
      render: (_: React.ReactNode, entity: any) => [
        <TableBtnWithPermission
          access="system.settings.ai_prompt.manager"
          key="edit"
          btnType="edit"
          onClick={() => openEdit(entity as PromptTemplate)}
          btnTitle={$t('编辑')}
        />,
        <TableBtnWithPermission
          access="system.settings.ai_prompt.manager"
          key="delete"
          btnType="delete"
          onClick={() => handleDelete(entity.id)}
          btnTitle={$t('删除')}
        />
      ]
    }
  ]

  return (
    <InsidePage
      pageTitle={$t('提示词管理')}
      description={$t('集中管理可复用的提示词模板，支持变量与标签，供 AI 调用与提示词工程复用。')}
      showBorder={false}
      scrollPage={false}
    >
      <div className="h-[calc(100%-1rem-36px)] pr-PAGE_INSIDE_X">
        <PageList
          ref={pageListRef}
          rowKey="id"
          afterNewBtn={[
            <WithPermission key="addPrompt" access="system.settings.ai_prompt.manager">
              <Button type="primary" key="addPromptBtn" onClick={openCreate}>
                {$t('添加提示词')}
              </Button>
            </WithPermission>
          ]}
          request={() => requestTemplates()}
          onSearchWordChange={(e) => setSearchWord(e.target.value)}
          showPagination={true}
          searchPlaceholder={$t('搜索名称或内容')}
          columns={columns}
        />
      </div>

      {modalOpen && (
        <PromptEditModalHost
          template={editing}
          onClose={() => setModalOpen(false)}
          onSaved={() => {
            setModalOpen(false)
            pageListRef.current?.reload()
          }}
        />
      )}
    </InsidePage>
  )
}

// PromptEditModalHost wraps the antd modal + the editor form so the ref-based
// save pattern stays isolated from the list page.
const PromptEditModalHost = ({
  template,
  onClose,
  onSaved
}: {
  template: PromptTemplate | null
  onClose: () => void
  onSaved: () => void
}) => {
  const { modal } = App.useApp()
  const editRef = useRef<PromptManagementHandle>(null)

  useEffect(() => {
    modal.confirm({
      title: template ? $t('编辑提示词') : $t('添加提示词'),
      content: <PromptEditorModal ref={editRef} template={template} />,
      width: 720,
      closable: true,
      onOk: () => {
        return editRef.current?.save().then((res) => {
          if (res === true) {
            onSaved()
          }
        })
      },
      onCancel: onClose,
      wrapClassName: 'ant-modal-without-footer',
      okText: $t('确认'),
      cancelText: $t('取消'),
      icon: <></>
    })
  }, [])

  return null
}

export default PromptManagement
