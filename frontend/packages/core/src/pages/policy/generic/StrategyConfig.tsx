import { LoadingOutlined } from "@ant-design/icons"
import InsidePage from "@common/components/aoplatform/InsidePage"
import WithPermission from "@common/components/aoplatform/WithPermission"
import { BasicResponse, STATUS_CODE, RESPONSE_TIPS, PLACEHOLDER } from "@common/const/const"
import { RouterParams } from "@common/const/type"
import { useGlobalContext } from "@common/contexts/GlobalStateContext"
import { useFetch } from "@common/hooks/http"
import { $t } from "@common/locales"
import { App, Button, Form, Input, InputNumber, Row, Select, Spin } from "antd"
import { forwardRef, useEffect, useImperativeHandle, useMemo, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import FilterTable from "../FilterTable"
import { PolicyOptions } from '@common/const/policy/consts'
import { Codebox } from '@common/components/postcat/api/Codebox'
import { useBreadcrumb } from "@common/contexts/BreadcrumbContext"

const StrategyConfig = forwardRef<{ save: () => void }, { driver: string }>((props, ref) => {
    const { driver } = props
    const { message } = App.useApp()
    const { teamId, serviceId, policyId } = useParams<RouterParams>();
    const [onEdit, setOnEdit] = useState<boolean>(!!teamId)
    const [form] = Form.useForm();
    const { fetchData } = useFetch()
    const { state } = useGlobalContext()
    const [loading, setLoading] = useState<boolean>(false)
    const navigator = useNavigate()
    const { setBreadcrumb } = useBreadcrumb()

    const backUrl = serviceId
        ? `/service/${teamId}/aiInside/${serviceId}/servicepolicy/datamasking/list`
        : '/globalpolicy/datamasking/list'

    useImperativeHandle(ref, () => ({
        save: onFinish
    }));

    const scope = serviceId === undefined ? 'global' : 'service'

    const getPolicyInfo = () => {
        setLoading(true)
        fetchData<BasicResponse<{ strategy: any }>>(`strategy/${scope}/${driver}`, { method: 'GET', eoParams: { team: teamId, service: serviceId, strategy: policyId } }).then(response => {
            const { code, data, msg } = response
            if (code === STATUS_CODE.SUCCESS) {
                setTimeout(() => {
                    form.setFieldsValue({
                        ...data.strategy,
                        type: driver,
                        config: data.strategy.config ? JSON.stringify(data.strategy.config, null, 2) : '{}',
                        filters: data.strategy.filters || []
                    })
                }, 0)
            } else {
                message.error(msg || $t(RESPONSE_TIPS.error))
            }
        }).finally(() => setLoading(false))
    };

    const onFinish: () => Promise<boolean | string> = () => {
        return form.validateFields().then((value) => {
            let config: any = {}
            try {
                config = value.config ? JSON.parse(value.config) : {}
            } catch (e) {
                message.error($t('配置内容不是合法的 JSON'))
                return Promise.reject($t('配置内容不是合法的 JSON'))
            }
            const body = {
                name: value.name,
                type: driver,
                priority: value.priority,
                desc: value.description,
                filters: value.filters || [],
                config
            }
            return fetchData<BasicResponse<{ service: { id: string } }>>(
                `strategy/${scope}/${driver}`,
                {
                    method: policyId === undefined ? 'POST' : 'PUT',
                    eoParams: { service: serviceId, team: teamId, strategy: policyId },
                    eoBody: body
                }).then(response => {
                    const { code, data, msg } = response
                    if (code === STATUS_CODE.SUCCESS) {
                        message.success(msg || $t(RESPONSE_TIPS.success))
                        navigator(backUrl)
                        return Promise.resolve(true)
                    } else {
                        message.error(msg || $t(RESPONSE_TIPS.error))
                        return Promise.reject(msg || $t(RESPONSE_TIPS.error))
                    }
                }).catch((errorInfo) => {
                    return Promise.reject(errorInfo)
                })
        })
    };

    useEffect(() => {
        if (policyId !== undefined) {
            setOnEdit(true);
            getPolicyInfo();
        } else {
            setOnEdit(false);
            form.setFieldValue('type', driver)
            form.setFieldValue('config', '{}')
        }
        return (form.setFieldsValue({}))
    }, [policyId, driver]);

    const policyOptions = useMemo(() => PolicyOptions.map((x) => ({ ...x, label: $t(x.label) })), [state.language])

    return (
        <InsidePage pageTitle={serviceId ? undefined : $t('编辑策略')}
            showBorder={false}
            scrollPage={false}
            className="overflow-y-auto"
            backUrl={backUrl}
        >
            <Spin indicator={<LoadingOutlined style={{ fontSize: 24 }} spin />} spinning={loading} wrapperClassName=' pb-PAGE_INSIDE_B pr-PAGE_INSIDE_X'>
                <WithPermission access={onEdit ? [`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.edit`] : ''}>
                    <Form
                        layout='vertical'
                        labelAlign='left'
                        scrollToFirstError
                        form={form}
                        className="w-full "
                        name="systemConfig"
                        onFinish={onFinish}
                        autoComplete="off"
                    >
                        <div>
                            <Form.Item
                                label={$t("策略名称")}
                                name="name"
                                rules={[{ required: true, whitespace: true }]}
                            >
                                <Input className="w-INPUT_NORMAL" placeholder={$t(PLACEHOLDER.input)} />
                            </Form.Item>

                            <Form.Item
                                label={$t("策略类型")}
                                name="type"
                                rules={[{ required: true }]}
                            >
                                <Select
                                    disabled
                                    showSearch
                                    optionFilterProp="label"
                                    className="w-INPUT_NORMAL"
                                    placeholder={$t(PLACEHOLDER.input)}
                                    options={policyOptions}
                                ></Select>
                            </Form.Item>

                            <Form.Item
                                label={$t("优先级")}
                                name={'priority'}
                                rules={[{ required: true }]}
                            >
                                <InputNumber className="w-INPUT_NORMAL" min={1} placeholder={$t(PLACEHOLDER.input)} />
                            </Form.Item>

                            <Form.Item
                                label={$t("描述")}
                                name="description"
                            >
                                <Input.TextArea className="w-INPUT_NORMAL" placeholder={$t(PLACEHOLDER.input)} />
                            </Form.Item>

                            <Form.Item
                                label={$t("匹配条件")}
                                name="filters"
                            >
                                <FilterTable />
                            </Form.Item>

                            <Form.Item
                                label={$t("配置")}
                                name="config"
                                rules={[{ required: true, message: $t('请填写配置') }]}
                            >
                                <Codebox
                                    editorTheme="vs-dark"
                                    width="100%"
                                    height="320px"
                                    language="json"
                                    enableToolbar={false}
                                />
                            </Form.Item>

                            <Row className="mb-[10px]">
                                <WithPermission access={onEdit ? [`${serviceId === undefined ? 'system.devops' : 'team.service'}.policy.edit`] : ''}>
                                    <Button type="primary" htmlType="submit">
                                        {$t('保存')}
                                    </Button>
                                </WithPermission>
                                <Button className="ml-btnrbase" type="default" onClick={() => navigator(backUrl)}>
                                    {$t('取消')}
                                </Button>
                            </Row>
                        </div>
                    </Form>
                </WithPermission>
            </Spin>
        </InsidePage>
    )
})
export default StrategyConfig
