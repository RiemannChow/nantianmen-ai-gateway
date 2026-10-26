import { ActionType } from "@ant-design/pro-components";
import { useEffect, useMemo, useRef, useState } from "react";
import { App, Button, message, Modal, Alert, Switch } from 'antd'
import PageList, { PageProColumns } from "@common/components/aoplatform/PageList";
import { $t } from "@common/locales";
import { useGlobalContext } from "@common/contexts/GlobalStateContext";
import { BasicResponse, DELETE_TIPS, RESPONSE_TIPS, STATUS_CODE } from "@common/const/const.tsx";
import { useFetch } from "@common/hooks/http";
import WithPermission from "@common/components/aoplatform/WithPermission.tsx";
import TableBtnWithPermission from "@common/components/aoplatform/TableBtnWithPermission";
import { useNavigate, useParams } from "react-router-dom";
import { RouterParams } from "@common/const/type";
import { PolicyPublishInfoType, PolicyPublishModalHandle } from "@common/const/type";
import { DrawerWithFooter } from "@common/components/aoplatform/DrawerWithFooter";
import { StrategyStatusEnum, StrategyStatusColorClass } from "@common/const/policy/consts";
import { checkAccess } from "@common/utils/permission";
import StrategyLogModal from "./StrategyLogModal";
import { StrategyPublishModalContent } from "./StrategyPublishModalContent";

const StrategyPolicy = (props: any) => {

  const {
    // 策略驱动：limiting / fuse / grey
    driver,
    // 是否显示发布按钮
    publishBtn = false,
    // 行操作
    rowOperation = []
  } = props;
  const { serviceId, teamId } = useParams<RouterParams>()
  const { state, accessData } = useGlobalContext()
  const navigator = useNavigate()
  const [drawerVisible, setDrawerVisible] = useState<boolean>(false)
  const [modalVisible, setModalVisible] = useState<boolean>(false);

  const [drawerData, setDrawerData] = useState<PolicyPublishInfoType>()
  const [isOkToPublish, setIsOkToPublish] = useState<boolean>(false)
  const drawerRef = useRef<PolicyPublishModalHandle>(null)
  const { modal } = App.useApp()

  /**
   * 列表ref
   */
  const pageListRef = useRef<ActionType>(null);

  /**
   * 请求数据
   */
  const { fetchData } = useFetch()

  /**
   * 搜索关键字
   */
  const [searchWord, setSearchWord] = useState<string>('')
  const [strategy, setStrategy] = useState<string>('')

  /**
   * 企业版能力（advanced_policy）授权状态
   */
  const [advancedEnabled, setAdvancedEnabled] = useState<boolean>(true)

  useEffect(() => {
    fetchData<BasicResponse<{ modules: { modules: { name: string; enabled: boolean }[] } }>>('license/modules', { method: 'GET' })
      .then((response) => {
        const { code, data } = response
        if (code === STATUS_CODE.SUCCESS) {
          const mod = (data.modules.modules || []).find((m: any) => m.name === 'advanced_policy')
          setAdvancedEnabled(!!(mod && mod.enabled))
        }
      })
      .catch(() => {/* 忽略，默认放行 */}
      )
  }, [driver])

  /**
   * 当前作用域
   */
  const scope = serviceId === undefined ? 'global' : 'service'

  /**
   * 获取列数据
   */
  const columns = useMemo(() => {
    const base: PageProColumns<any>[] = [
      {
        title: ('策略名称'),
        dataIndex: 'name',
        ellipsis: true,
        width: 160
      },
      {
        title: ('优先级'),
        dataIndex: 'priority',
        width: 90,
        ellipsis: true,
        sorter: (a: any, b: any) => (a.priority as number) - (b.priority as number)
      },
      {
        title: ('发布状态'),
        dataIndex: 'publishStatus',
        filters: true,
        onFilter: false,
        ellipsis: true,
        width: 110,
        valueEnum: new Map(
          Object.keys(StrategyStatusEnum).map(key =>
            [key,
            <span className={StrategyStatusColorClass[key as keyof typeof StrategyStatusColorClass]}>{$t(StrategyStatusEnum[key as keyof typeof StrategyStatusEnum])}</span>
            ])
        )
      },
      {
        title: ('启用'),
        dataIndex: 'isStop',
        filters: true,
        onFilter: false,
        ellipsis: true,
        width: 90,
        valueEnum: {
          false: { text: <span className="text-status_success">{$t('启用')}</span> },
          true: { text: <span className="text-status_fail">{$t('禁用')}</span> }
        }
      },
      {
        title: ('筛选条件'),
        dataIndex: 'filters',
        ellipsis: true
      },
      {
        title: ('更新者'),
        dataIndex: ['updater', 'name'],
        width: 140,
        ellipsis: true
      },
      {
        title: ('更新时间'),
        dataIndex: 'updateTime',
        width: 182,
        ellipsis: true
      }
    ]
    const res = base.map(x => {
      if (x.dataIndex === 'isStop') {
        x.render = (text: any, record: any) => <WithPermission access={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.edit`} ><AppSwitch checked={!record.isStop} record={record} onChange={changeOpenApiStatus} /></WithPermission>
      }
      return {
        ...x,
        title: (<span title={$t(x.title as string)}>{$t(x.title as string)}</span>)
      }
    })
    return res
  }, [state.language])

  /**
   * 操作列
   */
  const operation: PageProColumns<any>[] = rowOperation.length ? [
    {
      title: '',
      key: 'option',
      btnNums: rowOperation.length,
      fixed: 'right',
      valueType: 'option',
      render: (_: React.ReactNode, entity: any) => [
        ...(rowOperation.length && rowOperation.find((item: string) => item === 'edit') ? [<TableBtnWithPermission access={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.edit`} key="edit" btnType="edit" onClick={() => { openEditModal(entity) }} btnTitle="编辑" />] : []),
        ...(rowOperation.length && rowOperation.find((item: string) => item === 'logs') ? [<TableBtnWithPermission access={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.view`} key="logs" btnType="logs" onClick={() => { openLogsModal(entity) }} btnTitle="日志" />] : []),
        ...(rowOperation.length && rowOperation.find((item: string) => item === 'delete') ? [
          entity.isDelete ? <TableBtnWithPermission access={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.edit`} key="refresh" btnType="refresh" onClick={() => { restorePolicy(entity) }} btnTitle="恢复" /> :
            <TableBtnWithPermission access={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.delete`} key="delete" btnType="delete" onClick={() => { openModal('delete', entity) }} btnTitle="删除" />
        ] : []),
      ],
    }
  ] : []
  const handleCloseModal = () => {
    setModalVisible(false);
  };

  /**
   * 手动刷新表格数据
   */
  const manualReloadTable = () => {
    pageListRef.current?.reload()
  };

  /**
   * 更改启动状态
   * @param enabled 状态
   * @param entity 行数据
   */
  const changeOpenApiStatus = (enabled: boolean, entity: any) => {
    fetchData<BasicResponse<null>>(
      `strategy/${scope}/${driver}/${enabled ? 'enable' : 'disable'}`,
      {
        method: 'PATCH',
        eoParams: {
          service: serviceId,
          team: teamId,
          strategy: entity.id
        }
      }
    ).then(response => {
      const { code, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        message.success(msg || $t(RESPONSE_TIPS.success))
        manualReloadTable()
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
      }
    })
  }

  /**
   * 获取列表数据
   * @returns
   */
  const getPolicyList = (params: any & {
    pageSize: number;
    current: number;
  },
    sort: Record<string, string>,
    filter: Record<string, string>) => {

    return fetchData<BasicResponse<{ list: any[], total: number }>>(
      `strategy/${scope}/${driver}/list`,
      {
        method: 'GET',
        eoParams: {
          order: Object.keys(sort)?.[0],
          sort: Object.keys(sort)?.length > 0 ? Object.values(sort)?.[0] === 'descend' ? 'desc' : 'asc' : undefined,
          keyword: searchWord,
          service: serviceId,
          team: teamId,
        },
        eoTransformKeys: ['is_stop', 'is_delete', 'update_time', 'publish_status']
      }
    ).then(response => {
      const { code, data, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        return {
          data: data.list,
          total: data.total,
          success: true
        }
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
        return { data: [], success: false }
      }
    }).catch(() => {
      return { data: [], success: false }
    })
  }

  /**
   * 添加策略
   */
  const addPolicy = () => {
    if (serviceId === undefined) {
      navigator(`/globalpolicy/${driver}/create`)
    } else {
      navigator(`/service/${teamId}/aiInside/${serviceId}/servicepolicy/${driver}/create`)
    }
  }

  /**
   * 发布策略（仅全局策略）
   */
  const publish = async () => {
    message.loading($t(RESPONSE_TIPS.loading));
    const { code, data, msg } = await fetchData<BasicResponse<PolicyPublishInfoType>>(
      `strategy/global/${driver}/to-publishs`,
      { method: 'GET', eoTransformKeys: ['opt_time', 'is_publish', 'version_name', 'unpublish_msg'] }
    );
    message.destroy();
    if (code === STATUS_CODE.SUCCESS) {
      setDrawerVisible(true)
      setDrawerData(data)
      setIsOkToPublish(data.isPublish ?? true)
    } else {
      message.error(msg || $t(RESPONSE_TIPS.error));
      return
    }
  }

  /**
   * 编辑
   */
  const openEditModal = (entity: any) => {
    if (serviceId === undefined) {
      navigator(`/globalpolicy/${driver}/${entity.id}`)
    } else {
      navigator(`/service/${teamId}/aiInside/${serviceId}/servicepolicy/${driver}/${entity.id}`)
    }
  }

  /**
   * 日志
   * @param entity
   */
  const openLogsModal = (entity: any) => {
    setStrategy(entity.id)
    setModalVisible(true)
  }

  const openModal = async (type: 'delete', entity?: any) => {
    let title: string = ''
    let content: string | React.ReactNode = ''
    switch (type) {
      case 'delete':
        title = $t('删除')
        content = $t(DELETE_TIPS.default)
        break;
    }

    modal.confirm({
      title,
      content,
      onOk: () => {
        switch (type) {
          case 'delete':
            return deletePolicy(entity!).then((res) => { if (res === true) manualReloadTable() })
        }
      },
      width: 600,
      okText: $t('确认'),
      okButtonProps: {
        disabled: !checkAccess(`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.edit`, accessData)
      },
      cancelText: $t('取消'),
      closable: true,
      icon: <></>,
    })
  }

  /**
   * 删除
   * @param entity
   */
  const deletePolicy = (entity: any) => {
    return fetchData<BasicResponse<null>>(
      `strategy/${scope}/${driver}`,
      {
        method: 'DELETE',
        eoParams: {
          service: serviceId,
          team: teamId,
          strategy: entity.id
        },
      }
    ).then(response => {
      const { code, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        message.success(msg || $t(RESPONSE_TIPS.success))
        return Promise.resolve(true)
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
        return Promise.reject(msg || $t(RESPONSE_TIPS.error))
      }
    }).catch((errorInfo) => Promise.reject(errorInfo))
  }

  /**
   * 恢复
   * @param entity
   */
  const restorePolicy = (entity: any) => {
    fetchData<BasicResponse<null>>(
      `strategy/${scope}/${driver}/restore`,
      {
        method: 'PATCH',
        eoParams: {
          service: serviceId,
          team: teamId,
          strategy: entity.id
        },
      }
    ).then(response => {
      const { code, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        message.success(msg || $t(RESPONSE_TIPS.success))
        manualReloadTable()
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
      }
    })
  }

  const onSubmit = () => {
    return drawerRef.current?.publish()?.then((res) => {
      manualReloadTable();
      return res;
    });
  }

  return (
    <>
      {!advancedEnabled && (
        <Alert
          className="mb-[12px]"
          type="warning"
          showIcon
          message={$t('该能力属于企业版（高级策略）')}
          description={$t('当前 License 未包含 advanced_policy 模块，无法新建或发布该策略。请升级企业版后使用。')}
        />
      )}
      <PageList<any>
        id={`strategy_${driver}_list`}
        ref={pageListRef}
        columns={[...columns, ...operation]}
        request={async (params: any & {
          pageSize: number;
          current: number;
        },
          sort: Record<string, string>,
          filter: Record<string, string>) => getPolicyList(params, sort, filter)}
        addNewBtnTitle={$t("添加策略")}
        addNewBtnAccess={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.edit`}
        addNewBtnDisabled={!advancedEnabled}
        onAddNewBtnClick={() => { addPolicy() }}
        searchPlaceholder={$t("输入名称、筛选条件查找")}
        afterNewBtn={
          publishBtn && advancedEnabled && [<WithPermission key="removeFromDepPermission" access={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.publish`} >
            <Button className="mr-btnbase" key="removeFromDep" onClick={() => publish()}>{$t('发布')}</Button>
          </WithPermission>]
        }
        onSearchWordChange={(e) => {
          setSearchWord(e.target.value)
        }}
        manualReloadTable={manualReloadTable}
      />
      <DrawerWithFooter
        destroyOnClose={true}
        title={$t('申请发布')}
        width={'60%'}
        onClose={() => { setDrawerVisible(false) }}
        okBtnTitle={$t('发布')}
        open={drawerVisible}
        submitDisabled={!isOkToPublish}
        submitAccess={`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.publish`}
        onSubmit={onSubmit}
      >
        <StrategyPublishModalContent
          ref={drawerRef}
          driver={driver}
          data={drawerData!}
        />
      </DrawerWithFooter>
      <Modal
        title={$t('处理日志')}
        visible={modalVisible}
        destroyOnClose={true}
        onCancel={handleCloseModal}
        footer={null}
        wrapClassName="modal-without-footer"
        width={1100}
        maskClosable={true}
      >
        <div className="pb-btnybase flex flex-nowrap flex-col h-full w-full items-center justify-between">
          <StrategyLogModal driver={driver} strategy={strategy}></StrategyLogModal>
        </div>
      </Modal>
    </>
  )
}

/**
 * 启用状态开关（内部组件，便于在 render 中携带 record）
 */
const AppSwitch = (props: { checked: boolean; record: any; onChange: (enabled: boolean, entity: any) => void }) => {
  return <Switch checked={props.checked} onChange={(e) => { props.onChange(e, props.record) }} />
}

export default StrategyPolicy;
