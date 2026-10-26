import InsidePage from '@common/components/aoplatform/InsidePage'
import { BasicResponse, RESPONSE_TIPS, STATUS_CODE } from '@common/const/const'
import { useFetch } from '@common/hooks/http'
import { $t } from '@common/locales/index.ts'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  Row,
  Statistic,
  Switch,
  Table,
  Typography
} from 'antd'
import { useEffect, useState } from 'react'
import {
  SemanticCacheConfig,
  SemanticCacheEntry,
  SemanticCacheStats
} from './type'

const SemanticCache = () => {
  const { message } = App.useApp()
  const { fetchData } = useFetch()
  const [config, setConfig] = useState<SemanticCacheConfig>({
    enabled: false,
    ttl_seconds: 604800,
    embedding_endpoint: '',
    embedding_key: ''
  })
  const [stats, setStats] = useState<SemanticCacheStats>({
    hits: 0,
    misses: 0,
    hit_rate: 0,
    entries: 0
  })
  const [entries, setEntries] = useState<SemanticCacheEntry[]>([])
  const [form] = Form.useForm()
  const [saving, setSaving] = useState(false)
  const [clearing, setClearing] = useState(false)

  // test panel state
  const [testModel, setTestModel] = useState('gpt-4')
  const [testPrompt, setTestPrompt] = useState('')
  const [testResult, setTestResult] = useState<string>('')
  const [testing, setTesting] = useState(false)

  const loadConfig = () => {
    fetchData<BasicResponse<{ config: SemanticCacheConfig }>>('ai/cache/config', { method: 'GET' })
      .then((res) => {
        if (res.code === STATUS_CODE.SUCCESS && res.data?.config) {
          setConfig(res.data.config)
          form.setFieldsValue(res.data.config)
        }
      })
      .catch(() => {})
  }

  const loadStats = () => {
    fetchData<BasicResponse<{ stats: SemanticCacheStats }>>('ai/cache/stats', { method: 'GET' })
      .then((res) => {
        if (res.code === STATUS_CODE.SUCCESS && res.data?.stats) setStats(res.data.stats)
      })
      .catch(() => {})
  }

  const loadEntries = () => {
    fetchData<BasicResponse<{ list: SemanticCacheEntry[] }>>('ai/cache/entries', { method: 'GET' })
      .then((res) => {
        if (res.code === STATUS_CODE.SUCCESS && res.data?.list) setEntries(res.data.list)
      })
      .catch(() => {})
  }

  useEffect(() => {
    loadConfig()
    loadStats()
    loadEntries()
  }, [])

  const saveConfig = () => {
    form
      .validateFields()
      .then((values) => {
        setSaving(true)
        fetchData<BasicResponse<null>>('ai/cache/config', {
          method: 'PUT',
          eoBody: {
            enabled: !!values.enabled,
            ttl_seconds: Number(values.ttl_seconds) || 604800,
            embedding_endpoint: values.embedding_endpoint || '',
            embedding_key: values.embedding_key || ''
          }
        })
          .then((res) => {
            if (res.code === STATUS_CODE.SUCCESS) {
              message.success($t('保存成功'))
              setConfig({
                enabled: !!values.enabled,
                ttl_seconds: Number(values.ttl_seconds) || 604800,
                embedding_endpoint: values.embedding_endpoint || '',
                embedding_key: values.embedding_key || ''
              })
              loadStats()
            } else {
              message.error(res.msg || $t(RESPONSE_TIPS.error))
            }
          })
          .catch(() => message.error($t(RESPONSE_TIPS.error)))
          .finally(() => setSaving(false))
      })
      .catch(() => {})
  }

  const clearCache = () => {
    setClearing(true)
    fetchData<BasicResponse<null>>('ai/cache/clear', { method: 'POST' })
      .then((res) => {
        if (res.code === STATUS_CODE.SUCCESS) {
          message.success($t('已清空缓存'))
          loadStats()
          loadEntries()
        } else {
          message.error(res.msg || $t(RESPONSE_TIPS.error))
        }
      })
      .catch(() => message.error($t(RESPONSE_TIPS.error)))
      .finally(() => setClearing(false))
  }

  const runTest = (mode: 'query' | 'store') => {
    if (!testPrompt) {
      message.warning($t('请输入提示词'))
      return
    }
    setTesting(true)
    const call =
      mode === 'query'
        ? fetchData<BasicResponse<{ entry: SemanticCacheEntry | null }>>('ai/cache/query', {
            method: 'POST',
            eoBody: { model: testModel, prompt: testPrompt, tenant: '' }
          })
        : fetchData<BasicResponse<null>>('ai/cache/store', {
            method: 'POST',
            eoBody: {
              model: testModel,
              prompt: testPrompt,
              tenant: '',
              response: 'cached-response-' + Date.now(),
              token_count: 0
            }
          })
    call
      .then((res) => {
        if (res.code === STATUS_CODE.SUCCESS) {
          if (mode === 'query') {
            const entry = (res.data as any)?.entry
            setTestResult(entry ? $t('命中缓存') : $t('未命中（miss）'))
          } else {
            setTestResult($t('已写入缓存，再次查询应命中'))
          }
          loadStats()
          loadEntries()
        } else {
          message.error(res.msg || $t(RESPONSE_TIPS.error))
        }
      })
      .catch(() => message.error($t(RESPONSE_TIPS.error)))
      .finally(() => setTesting(false))
  }

  const columns = [
    { title: $t('模型'), dataIndex: 'model', key: 'model' },
    { title: $t('租户'), dataIndex: 'tenant', key: 'tenant', render: (v: string) => v || '-' },
    {
      title: $t('Prompt Hash'),
      dataIndex: 'prompt_hash',
      key: 'prompt_hash',
      ellipsis: true
    },
    { title: $t('Token 数'), dataIndex: 'token_count', key: 'token_count' },
    {
      title: $t('响应'),
      dataIndex: 'response',
      key: 'response',
      ellipsis: true,
      render: (v: string) => <Typography.Text>{v || '-'}</Typography.Text>
    }
  ]

  return (
    <InsidePage
      pageTitle={$t('语义缓存')}
      description={$t(
        '对相同或高度相似的提示词缓存响应，降低重复调用成本。可配置 TTL 与可选 embedding 端点（未配置时按归一化哈希精确命中）。'
      )}
      showBorder={false}
      scrollPage={false}
    >
      <Row gutter={16} className="mb-4">
        <Col span={6}>
          <Card>
            <Statistic title={$t('命中率')} value={(stats.hit_rate * 100).toFixed(1)} suffix="%" />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={$t('命中次数')} value={stats.hits} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={$t('未命中次数')} value={stats.misses} />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title={$t('缓存条目')} value={stats.entries} />
          </Card>
        </Col>
      </Row>

      <Card title={$t('缓存配置')} className="mb-4">
        <Form form={form} layout="vertical" initialValues={config}>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="enabled" label={$t('启用缓存')} valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="ttl_seconds" label={$t('TTL（秒）')} rules={[{ required: true }]}>
                <InputNumber min={60} style={{ width: '100%' }} placeholder="604800" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="embedding_endpoint" label={$t('Embedding 端点（可选）')}>
                <Input placeholder="https://api.openai.com/v1/embeddings" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="embedding_key" label={$t('Embedding Key（可选）')}>
            <Input.Password placeholder={$t('配置后启用向量近邻命中')} />
          </Form.Item>
          <Button type="primary" loading={saving} onClick={saveConfig}>
            {$t('保存配置')}
          </Button>
        </Form>
      </Card>

      <Card
        title={$t('缓存条目')}
        className="mb-4"
        extra={
          <Button danger loading={clearing} onClick={clearCache}>
            {$t('清空缓存')}
          </Button>
        }
      >
        <Table rowKey="key" columns={columns} dataSource={entries} pagination={{ pageSize: 10 }} />
      </Card>

      <Card title={$t('命中验证')}>
        <Alert
          className="mb-3"
          type="info"
          showIcon
          message={$t('演示：先「写入」一个提示词，再「查询」相同提示词即可看到命中。')}
        />
        <Row gutter={16} align="bottom">
          <Col span={6}>
            <Typography.Text>{$t('模型')}</Typography.Text>
            <Input value={testModel} onChange={(e) => setTestModel(e.target.value)} />
          </Col>
          <Col span={12}>
            <Typography.Text>{$t('提示词')}</Typography.Text>
            <Input value={testPrompt} onChange={(e) => setTestPrompt(e.target.value)} />
          </Col>
          <Col span={6}>
            <Button onClick={() => runTest('query')} loading={testing} className="mr-2">
              {$t('查询')}
            </Button>
            <Button type="primary" onClick={() => runTest('store')} loading={testing}>
              {$t('写入')}
            </Button>
          </Col>
        </Row>
        {testResult && (
          <Typography.Paragraph className="mt-3">
            {$t('结果')}：<Typography.Text strong>{testResult}</Typography.Text>
          </Typography.Paragraph>
        )}
      </Card>
    </InsidePage>
  )
}

export default SemanticCache
