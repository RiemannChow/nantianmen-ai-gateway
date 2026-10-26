import InsidePage from '@common/components/aoplatform/InsidePage'
import WithPermission from '@common/components/aoplatform/WithPermission'
import { BasicResponse, PLACEHOLDER, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { useGlobalContext } from '@common/contexts/GlobalStateContext'
import { useFetch } from '@common/hooks/http'
import { $t } from '@common/locales'
import { App, Alert, Button, Form, Input, Row, Select, Switch } from 'antd'
import { useEffect, useState } from 'react'

type FieldDef = {
  key: string
  label: string
  type: 'text' | 'password' | 'switch'
  required?: boolean
  extra?: string
  placeholder?: string
}

// 各第三方登录驱动的配置字段。新增驱动只需在此补充字段描述，
// 页面会根据所选驱动动态渲染对应的表单。
const DRIVER_FIELDS: Record<string, FieldDef[]> = {
  feishu: [
    { key: 'clientId', label: 'APP ID', type: 'text', required: true, extra: 'APP ID 参数位于飞书开发人员控制台中的应用程序凭证和基础信息页面上' },
    { key: 'clientSecret', label: 'APP Secret', type: 'password', required: true, extra: 'APP Secret 参数位于飞书开发人员控制台中的应用程序凭证和基础信息页面上' }
  ],
  ldap: [
    { key: 'host', label: 'LDAP 服务器地址', type: 'text', required: true, extra: '例如 ldap://192.168.1.10:389 或 ldaps://192.168.1.10:636', placeholder: 'ldap://host:389' },
    { key: 'useSsl', label: '使用 SSL/TLS（StartTLS）', type: 'switch' },
    { key: 'bindDn', label: '绑定 DN', type: 'text', required: true, extra: '用于连接 LDAP 目录的服务账号 DN（匿名绑定可留空）', placeholder: 'cn=admin,dc=example,dc=com' },
    { key: 'bindPassword', label: '绑定密码', type: 'password', required: true },
    { key: 'baseDn', label: '搜索基准 DN', type: 'text', required: true, placeholder: 'ou=users,dc=example,dc=com' },
    { key: 'userFilter', label: '用户过滤条件', type: 'text', extra: '占位符 %s 会被登录用户名替换', placeholder: '(&(objectClass=person)(uid=%s))' },
    { key: 'userDnAttr', label: '用户 DN 属性', type: 'text', extra: '用于构建用户 DN 的属性名', placeholder: 'uid' },
    { key: 'attrUsername', label: '用户名属性', type: 'text', placeholder: 'uid' },
    { key: 'attrName', label: '姓名属性', type: 'text', placeholder: 'cn' },
    { key: 'attrEmail', label: '邮箱属性', type: 'text', placeholder: 'mail' },
    { key: 'attrMobile', label: '手机属性', type: 'text', placeholder: 'mobile' }
  ]
}

type AuthSetting = {
  config: Record<string, string>
  enable: boolean
}

const camelToSnake = (s: string): string => s.replace(/[A-Z]/g, (m) => '_' + m.toLowerCase())

const Auth = () => {
  const { message } = App.useApp()
  const [form] = Form.useForm()
  const { fetchData } = useFetch()
  const { state } = useGlobalContext()
  const [thirdPartyDrivers, setThirdPartyDrivers] = useState<{ label: string; value: string }[]>([])
  // LDAP / SSO 企业版能力授权状态
  const [ssoEnabled, setSsoEnabled] = useState<boolean>(true)

  useEffect(() => {
    fetchData<BasicResponse<{ modules: { modules: { name: string; enabled: boolean }[] } }>>('license/modules', { method: 'GET' })
      .then((response) => {
        const { code, data } = response
        if (code === STATUS_CODE.SUCCESS) {
          const mod = (data.modules.modules || []).find((m: any) => m.name === 'sso')
          setSsoEnabled(!!(mod && mod.enabled))
        }
      })
      .catch(() => { /* 忽略，默认放行 */ })
  }, [])

  const getThirdPartyAuthList = () => {
    fetchData<
      BasicResponse<{
        drivers: {
          name: string
          value: string
        }[]
      }>
    >('account/third', { method: 'GET' }).then((response) => {
      const { code, data, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        let drivers = data.drivers.map((item: any) => ({ label: item.name, value: item.value }))
        // 未授权 SSO 企业版能力时，隐藏 LDAP 配置入口
        if (!ssoEnabled) {
          drivers = drivers.filter((d) => d.value !== 'ldap')
        }
        setThirdPartyDrivers(drivers)
        if (drivers.length) {
          form.setFieldValue('authType', drivers[0].value)
          getThirdPartyAuthSetting()
        }
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
      }
    })
  }

  const getThirdPartyAuthSetting = () => {
    const authType = form.getFieldValue('authType')
    if (!authType) return
    fetchData<BasicResponse<{ driver: AuthSetting }>>(`account/third/${authType}`, {
      method: 'GET',
      eoTransformKeys: ['client_id', 'client_secret', 'use_ssl', 'bind_dn', 'bind_password', 'base_dn', 'user_filter', 'user_dn_attr', 'attr_username', 'attr_name', 'attr_email', 'attr_mobile', 'host']
    }).then((response) => {
      const { code, data, msg } = response
      if (code === STATUS_CODE.SUCCESS) {
        const config = data.driver?.config || {}
        const values: Record<string, any> = { enabled: data.driver?.enable || false }
        ;(DRIVER_FIELDS[authType] || []).forEach((f) => {
          values[f.key] = config[f.key] ?? (f.type === 'switch' ? false : '')
        })
        form.setFieldsValue(values)
      } else {
        message.error(msg || $t(RESPONSE_TIPS.error))
      }
    })
  }

  const onFinish = () => {
    form.validateFields().then((value) => {
      const authType = value.authType
      const config: Record<string, any> = {}
      ;(DRIVER_FIELDS[authType] || []).forEach((f) => {
        if (value[f.key] !== undefined && value[f.key] !== '') {
          config[camelToSnake(f.key)] = value[f.key]
        }
      })
      return fetchData<BasicResponse<null>>(`account/third/${authType}`, {
        method: 'POST',
        eoBody: {
          enable: value.enabled,
          config
        }
      })
        .then((response) => {
          const { code, msg } = response
          if (code === STATUS_CODE.SUCCESS) {
            message.success(msg || $t(RESPONSE_TIPS.success))
            return Promise.resolve(true)
          } else {
            message.error(msg || $t(RESPONSE_TIPS.error))
            return Promise.reject(msg || $t(RESPONSE_TIPS.error))
          }
        })
        .catch((errorInfo) => {
          return Promise.reject(errorInfo)
        })
    })
  }

  const currentFields = DRIVER_FIELDS[form.getFieldValue('authType')] || []
  const isLdap = form.getFieldValue('authType') === 'ldap'

  return (
    <InsidePage pageTitle={$t('鉴权')} showBorder={false} contentClassName="pr-PAGE_INSIDE_X" scrollPage={false} description={$t('系统用户账号登录授权配置')}>
      {isLdap && !ssoEnabled && (
        <Alert
          className="mb-[12px]"
          type="warning"
          showIcon
          message={$t('该能力属于企业版（SSO / LDAP 单点登录）')}
          description={$t('当前 License 未包含 sso 模块，无法启用 LDAP 登录。请升级企业版后使用。')}
        />
      )}
      <WithPermission access="">
        <Form
          layout="vertical"
          labelAlign="left"
          scrollToFirstError
          form={form}
          className="mx-auto"
          name="authConfig"
          onFinish={onFinish}
          autoComplete="off"
        >
          <Form.Item
            label={$t('授权类型')}
            name="authType"
            rules={[{ required: true, message: $t('请选择授权类型') }]}
          >
            <Select className="w-INPUT_NORMAL" placeholder={$t('请选择授权类型')} onChange={getThirdPartyAuthSetting} options={thirdPartyDrivers} />
          </Form.Item>

          {currentFields.map((f) => (
            <Form.Item
              key={f.key}
              label={$t(f.label)}
              name={f.key}
              rules={f.required ? [{ required: true, whitespace: true, message: $t(`请输入${f.label}`) }] : undefined}
              extra={f.extra ? $t(f.extra) : undefined}
            >
              {f.type === 'switch' ? (
                <Switch checkedChildren={$t('启用')} unCheckedChildren={$t('停用')} disabled={isLdap && !ssoEnabled} />
              ) : f.type === 'password' ? (
                <Input.Password className="w-INPUT_NORMAL" placeholder={$t(PLACEHOLDER.input)} autoComplete="new-password" disabled={isLdap && !ssoEnabled} />
              ) : (
                <Input className="w-INPUT_NORMAL" placeholder={$t(PLACEHOLDER.input)} autoComplete="off" disabled={isLdap && !ssoEnabled} />
              )}
            </Form.Item>
          ))}

          <Form.Item label={$t('启用授权')} name="enabled" valuePropName="checked">
            <Switch checkedChildren={$t('启用')} unCheckedChildren={$t('停用')} disabled={isLdap && !ssoEnabled} />
          </Form.Item>

          <Row className="mb-[10px]">
            <WithPermission access="system.devops.system_setting.edit">
              <Button type="primary" htmlType="submit" disabled={isLdap && !ssoEnabled}>
                {$t('保存')}
              </Button>
            </WithPermission>
          </Row>
        </Form>
      </WithPermission>
    </InsidePage>
  )
}

export default Auth
