import { $t } from "@common/locales/index.ts";
import DataMasking from "./dataMasking/DataMasking";
import StrategyPolicy from "./generic/StrategyPolicy";
import PolicyTabContainer from "./policyTabContainer.tsx";

const servicePolicy = () => {
  /**
   * tab列表
   */
  const tabItems = [
    {
      key: 'dataMasking',
      label: $t('数据脱敏'),
      children: <div className="pr-[40px] h-full preview-document mb-PAGE_INSIDE_B"><DataMasking publishBtn={false} rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'limiting',
      label: $t('精细限流'),
      children: <div className="pr-[40px] h-full preview-document mb-PAGE_INSIDE_B"><StrategyPolicy driver="limiting" publishBtn={false} rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'fuse',
      label: $t('熔断'),
      children: <div className="pr-[40px] h-full preview-document mb-PAGE_INSIDE_B"><StrategyPolicy driver="fuse" publishBtn={false} rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'grey',
      label: $t('灰度发布'),
      children: <div className="pr-[40px] h-full preview-document mb-PAGE_INSIDE_B"><StrategyPolicy driver="grey" publishBtn={false} rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'piiMask',
      label: $t('PII 脱敏'),
      children: <div className="pr-[40px] h-full preview-document mb-PAGE_INSIDE_B"><StrategyPolicy driver="pii-mask" publishBtn={false} rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'promptInjection',
      label: $t('注入防护'),
      children: <div className="pr-[40px] h-full preview-document mb-PAGE_INSIDE_B"><StrategyPolicy driver="prompt-injection" publishBtn={false} rowOperation={['edit', 'logs', 'delete']} /></div>
    },
    {
      key: 'contentModeration',
      label: $t('内容审核'),
      children: <div className="pr-[40px] h-full preview-document mb-PAGE_INSIDE_B"><StrategyPolicy driver="content-moderation" publishBtn={false} rowOperation={['edit', 'logs', 'delete']} /></div>
    }
  ]

  return (
    <>
      <PolicyTabContainer tabs={tabItems} />
    </>
  )
}

export default servicePolicy;